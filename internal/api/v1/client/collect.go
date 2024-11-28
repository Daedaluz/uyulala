package client

import (
	"database/sql"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"
	"uyulala/internal/api"
	"uyulala/internal/api/application"
	"uyulala/internal/db"
	"uyulala/internal/db/appdb"
	"uyulala/internal/db/challengedb"
	"uyulala/internal/db/keydb"
	"uyulala/internal/db/sessiondb"
	"uyulala/internal/db/userdb"
	"uyulala/openid/discovery"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/spf13/viper"
)

type SignatureData struct {
	Text string `json:"text"`
	Data []byte `json:"data"`
}

type CollectResponseExp struct {
	ChallengeID          string                                  `json:"challengeId"`
	UserID               string                                  `json:"userId"`
	SignatureData        SignatureData                           `json:"signatureData"`
	PublicData           *protocol.CredentialAssertion           `json:"assertion"`
	AssertionSignature   *protocol.ParsedCredentialAssertionData `json:"assertionSignature,omitempty"`
	AttestationSignature *protocol.ParsedCredentialCreationData  `json:"attestationSignature,omitempty"`
	Credential           *webauthn.Credential                    `json:"credential"`
	Signed               time.Time                               `json:"signed"`
	Status               string                                  `json:"status"`
}

type CollectResponse struct {
	ChallengeID         string                                     `json:"challengeId"`
	UserID              string                                     `json:"userId"`
	Status              string                                     `json:"status"`
	Signed              time.Time                                  `json:"signed"`
	UserPresent         bool                                       `json:"userPresent"`
	UserVerified        bool                                       `json:"userVerified"`
	PublicKey           []byte                                     `json:"publicKey"`
	AssertionResponse   *protocol.AuthenticatorAssertionResponse   `json:"assertionResponse,omitempty"`
	AttestationResponse *protocol.AuthenticatorAttestationResponse `json:"attestationResponse,omitempty"`
	Challenge           protocol.URLEncodedBase64                  `json:"challenge"`
	SignatureData       SignatureData                              `json:"signatureData"`
}

func (c *CollectResponseExp) Response() *CollectResponse {
	res := &CollectResponse{
		ChallengeID:   c.ChallengeID,
		UserID:        c.UserID,
		Status:        c.Status,
		Signed:        c.Signed,
		PublicKey:     c.Credential.PublicKey,
		Challenge:     c.PublicData.Response.Challenge,
		SignatureData: c.SignatureData,
	}
	if c.AttestationSignature != nil {
		res.AttestationResponse = &c.AttestationSignature.Raw.AttestationResponse
	}
	if c.AssertionSignature != nil {
		res.AssertionResponse = &c.AssertionSignature.Raw.AssertionResponse
		res.UserPresent = c.AssertionSignature.Response.AuthenticatorData.Flags.UserPresent()
		res.UserVerified = c.AssertionSignature.Response.AuthenticatorData.Flags.UserVerified()
	}
	return res
}

func collectResponseFromChallenge(challenge *challengedb.Data) *CollectResponseExp {
	res := &CollectResponseExp{
		ChallengeID:   challenge.ID,
		SignatureData: SignatureData{Text: challenge.SignatureText, Data: challenge.SignatureData},
		Status:        challenge.Status,
		Signed:        challenge.Signed.Time,
	}
	_ = db.GobDecodeData(challenge.PubData, &res.PublicData)
	_ = db.GobDecodeData(challenge.Signature, &res.AssertionSignature)
	if res.AssertionSignature.Raw.AssertionResponse.ClientDataJSON == nil {
		res.AssertionSignature = nil
	}
	_ = db.GobDecodeData(challenge.Signature, &res.AttestationSignature)
	if res.AttestationSignature.Raw.AttestationResponse.ClientDataJSON == nil {
		res.AttestationSignature = nil
	}

	_ = db.GobDecodeData(challenge.Credential, &res.Credential)
	return res
}

func collectBIDFlow(context *gin.Context, app *appdb.Application) {
	in := gin.H{}
	if err := context.Bind(&in); err != nil {
		api.AbortError(context, http.StatusBadRequest, "invalid_request", "Invalid request", err)
		return
	}
	var challengeID string
	if challenge, ok := in["challengeId"].(string); ok {
		challengeID = challenge
	}
	if challengeID == "" {
		api.AbortError(context, http.StatusBadRequest, "invalid_request", "Missing challenge_id", nil)
		return
	}
	challenge, err := challengedb.GetChallenge(context, challengeID)
	if err != nil {
		api.AbortError(context, http.StatusInternalServerError, "invalid_challenge", "Unexpected error", err)
		return
	}
	if challenge.OAuth2Context != "" {
		api.AbortError(context, http.StatusBadRequest, "invalid_request", "This challenge was started with the oauth2 flow", nil)
		return
	}
	if challenge.AppID != app.ID {
		api.AbortError(context, http.StatusBadRequest, "invalid_request", "Challenge not intended for this client", nil)
		return
	}

	if !challenge.ValidateBIDCollect(context) {
		return
	}
	if err := challengedb.SetChallengeStatus(context, challengeID, challengedb.StatusCollected); err != nil {
		api.AbortError(context, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return
	}

	response := collectResponseFromChallenge(challenge)
	var id []byte
	if response.AssertionSignature != nil {
		id = response.AssertionSignature.RawID
	} else {
		id = response.AttestationSignature.Raw.RawID
	}
	key, err := userdb.GetKey(context, id)
	if err != nil {
		api.AbortError(context, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return
	}
	response.UserID = key.UserID
	context.JSON(http.StatusOK, response.Response())
}

func createIDToken(context *gin.Context, sessionID, userID, nonce string, app *appdb.Application, appKey jwk.Key,
	response *CollectResponseExp) (string, error) {
	lastAuth, err := userdb.GetAuthTime(context, userID, app.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		api.AbortError(context, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return "", err
	}

	startTime := time.Now()
	if response != nil {
		startTime = response.Signed
	}

	token := jwt.New()
	_ = token.Set("sub", userID)
	_ = token.Set("iss", viper.GetString("issuer"))
	_ = token.Set("aud", app.ID)
	_ = token.Set("exp", startTime.Add(viper.GetDuration("idToken.length")).Unix())
	_ = token.Set("nbf", startTime.Unix())
	_ = token.Set("auth_time", lastAuth.Unix())
	_ = token.Set("iat", time.Now().Unix())
	if response != nil {
		_ = token.Set("uv", response.AssertionSignature.Response.AuthenticatorData.Flags.UserVerified())
		_ = token.Set("up", response.AssertionSignature.Response.AuthenticatorData.Flags.UserPresent())
	}

	if sessionID != "" {
		_ = token.Set("sid", sessionID)
	}
	if nonce != "" {
		_ = token.Set("nonce", nonce)
	}

	data, err := jwt.Sign(token, jwa.SignatureAlgorithm(appKey.Algorithm()), appKey)
	if err != nil {
		api.AbortError(context, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return "", err
	}
	tokenString := string(data)
	return tokenString, nil
}

func createAccessToken(ctx *gin.Context, sessionID, userID string, key jwk.Key, app *appdb.Application,
	response *CollectResponseExp) (string, error) {
	startTime := time.Now()
	if response != nil {
		startTime = response.Signed
	}

	token := jwt.New()
	_ = token.Set("sub", userID)
	_ = token.Set("iss", viper.GetString("issuer"))
	_ = token.Set("aud", app.ID)
	_ = token.Set("exp", startTime.Add(viper.GetDuration("accessToken.length")).Unix())
	_ = token.Set("nbf", startTime.Unix())
	_ = token.Set("iat", time.Now().Unix())
	if sessionID != "" {
		_ = token.Set("sid", sessionID)
	}
	extra := viper.GetStringMap("accessToken.extension")
	for k, v := range extra {
		_ = token.Set(k, v)
	}
	hdrs := jws.NewHeaders()
	_ = hdrs.Set(jws.TypeKey, "at+jwt")
	data, err := jwt.Sign(token, jwa.SignatureAlgorithm(key.Algorithm()), key, jwt.WithJwsHeaders(hdrs))
	if err != nil {
		api.AbortError(ctx, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return "", err
	}
	tokenString := string(data)
	return tokenString, nil
}

type TokenResponse struct {
	AccessToken  string `json:"access_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty" `
}

func collectOAuth2Flow(context *gin.Context, app *appdb.Application) {
	var (
		idToken      string
		accessToken  string
		refreshToken string
		sessionID    string
		resultScopes []string
	)
	key, err := keydb.GetKey(context, app.KeyID)
	if err != nil {
		api.AbortError(context, http.StatusInternalServerError, "no_key", "This shouldn't happen. couldn't find the signing key.", err)
		return
	}
	appKey, err := key.GetPrivateJWK()
	if err != nil {
		api.AbortError(context, http.StatusInternalServerError, "key_parsing_error", "This shouldn't happen. couldn't parse the private key for signing", err)
		return
	}
	switch context.PostForm("grant_type") {
	case discovery.GrantTypeAuthorizationCode:
		challenge := application.GetCurrentChallenge(context)
		if !challenge.ValidateOAuthCollect(context) {
			return
		}
		if err := challengedb.SetChallengeStatus(context, challenge.ID, challengedb.StatusCollected); err != nil {
			api.AbortError(context, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
			return
		}
		response := collectResponseFromChallenge(challenge)
		oauth2Ctx, _ := url.ParseQuery(challenge.OAuth2Context)
		scopes := strings.FieldsFunc(oauth2Ctx.Get("scope"), func(c rune) bool {
			switch c {
			case ' ', '\t', '\r', '\n':
				return true
			}
			return false
		})
		userKey, err := userdb.GetKey(context, response.AssertionSignature.RawID)
		if err != nil {
			api.AbortError(context, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
			return
		}

		if slices.Contains(scopes, "offline_access") {
			sess, err := sessiondb.Create(context, userKey.UserID, app.ID, oauth2Ctx.Get("scope"))
			if err != nil {
				api.AbortError(context, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
				return
			}
			resultScopes = append(resultScopes, "offline_access")
			sessionID = sess.ID
			refreshToken, err = sess.CreateRefreshToken(appKey)
			if err != nil {
				api.AbortError(context, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
				return
			}
			accessToken, err = createAccessToken(context, sessionID, userKey.UserID, appKey, app, response)
			if err != nil {
				return
			}
		}

		if slices.Contains(scopes, "openid") {
			// TODO: add ACR data
			resultScopes = append(resultScopes, "openid")
			idToken, err = createIDToken(context, sessionID, userKey.UserID, oauth2Ctx.Get("nonce"), app, appKey, response)
			if err != nil {
				return
			}
		}

	case discovery.GrantTypeRefresh:
		session := application.GetCurrentSession(context)
		scopes := strings.FieldsFunc(session.RequestedScopes, func(c rune) bool {
			switch c {
			case ' ', '\t', '\r', '\n':
				return true
			}
			return false
		})

		if err := sessiondb.Rotate(context, session); err != nil {
			api.AbortError(context, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
			return
		} else {
			tmp, err := session.CreateRefreshToken(appKey)
			if err != nil {
				api.AbortError(context, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
				return
			}
			refreshToken = tmp
		}
		if slices.Contains(scopes, "openid") {
			idToken, err = createIDToken(context, session.ID, session.UserID, "", app, appKey, nil)
			resultScopes = append(resultScopes, "openid")
			if err != nil {
				api.AbortError(context, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
				return
			}
		}
		resultScopes = append(resultScopes, "offline_access")
		accessToken, err = createAccessToken(context, session.ID, session.UserID, appKey, app, nil)
		if err != nil {
			api.AbortError(context, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
			return
		}
	case discovery.GrantTypeCIBA:
		challenge := application.GetCurrentChallenge(context)
		if !challenge.ValidateOAuthCollect(context) {
			return
		}
		if err := challengedb.SetChallengeStatus(context, challenge.ID, challengedb.StatusCollected); err != nil {
			api.AbortError(context, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
			return
		}
		response := collectResponseFromChallenge(challenge)
		oauth2Ctx, _ := url.ParseQuery(challenge.OAuth2Context)
		scopes := strings.FieldsFunc(oauth2Ctx.Get("scope"), func(c rune) bool {
			switch c {
			case ' ', '\t', '\r', '\n':
				return true
			}
			return false
		})

		userKey, err := userdb.GetKey(context, response.AssertionSignature.RawID)
		if err != nil {
			api.AbortError(context, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
			return
		}

		if slices.Contains(scopes, "offline_access") {
			sess, err := sessiondb.Create(context, userKey.UserID, app.ID, oauth2Ctx.Get("scope"))
			if err != nil {
				api.AbortError(context, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
				return
			}
			resultScopes = append(resultScopes, "offline_access")
			sessionID = sess.ID
			refreshToken, err = sess.CreateRefreshToken(appKey)
			if err != nil {
				api.AbortError(context, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
				return
			}
			accessToken, err = createAccessToken(context, sessionID, userKey.UserID, appKey, app, response)
			if err != nil {
				return
			}
		}

		if slices.Contains(scopes, "openid") {
			// TODO: add ACR data
			resultScopes = append(resultScopes, "openid")
			idToken, err = createIDToken(context, sessionID, userKey.UserID, oauth2Ctx.Get("nonce"), app, appKey, response)
			if err != nil {
				return
			}
		}
	case "":
		api.AbortError(context, http.StatusBadRequest, "invalid_request", "Missing grant_type", nil)
	}
	context.JSON(http.StatusOK, &TokenResponse{
		AccessToken:  accessToken,
		Scope:        strings.Join(resultScopes, " "),
		IDToken:      idToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	})
}

func collectHandler(context *gin.Context) {
	app := application.GetCurrentApplication(context)
	if context.ContentType() == "application/json" {
		collectBIDFlow(context, app)
	} else if context.ContentType() == "application/x-www-form-urlencoded" {
		collectOAuth2Flow(context, app)
	} else {
		api.AbortError(context, http.StatusBadRequest, "invalid_content_type", "Invalid content type", nil)
	}
}
