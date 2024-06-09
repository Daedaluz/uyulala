package client

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/spf13/viper"
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
)

type SignatureData struct {
	Text string `json:"text"`
	Data []byte `json:"data"`
}

type CollectResponseExp struct {
	ChallengeID   string                                  `json:"challengeId"`
	UserID        string                                  `json:"userId"`
	SignatureData SignatureData                           `json:"signatureData"`
	PublicData    *protocol.CredentialAssertion           `json:"assertion"`
	Signature     *protocol.ParsedCredentialAssertionData `json:"signature"`
	Credential    *webauthn.Credential                    `json:"credential"`
	Signed        time.Time                               `json:"signed"`
	Status        string                                  `json:"status"`
}

type CollectResponse struct {
	ChallengeID   string                                  `json:"challengeId"`
	UserID        string                                  `json:"userId"`
	Status        string                                  `json:"status"`
	Signed        time.Time                               `json:"signed"`
	UserPresent   bool                                    `json:"userPresent"`
	UserVerified  bool                                    `json:"userVerified"`
	PublicKey     []byte                                  `json:"publicKey"`
	Response      protocol.AuthenticatorAssertionResponse `json:"assertionResponse"`
	Challenge     protocol.URLEncodedBase64               `json:"challenge"`
	SignatureData SignatureData                           `json:"signatureData"`
}

func (c *CollectResponseExp) Response() *CollectResponse {
	return &CollectResponse{
		ChallengeID:   c.ChallengeID,
		UserID:        c.UserID,
		Status:        c.Status,
		Signed:        c.Signed,
		UserPresent:   c.Signature.Response.AuthenticatorData.Flags.UserPresent(),
		UserVerified:  c.Signature.Response.AuthenticatorData.Flags.UserVerified(),
		PublicKey:     c.Credential.PublicKey,
		Response:      c.Signature.Raw.AssertionResponse,
		Challenge:     c.PublicData.Response.Challenge,
		SignatureData: c.SignatureData,
	}
}

func collectResponseFromChallenge(challenge *challengedb.Data) *CollectResponseExp {
	res := &CollectResponseExp{
		ChallengeID:   challenge.ID,
		SignatureData: SignatureData{Text: challenge.SignatureText, Data: challenge.SignatureData},
		Status:        challenge.Status,
		Signed:        challenge.Signed.Time,
	}
	_ = db.GobDecodeData(challenge.PubData, &res.PublicData)
	_ = db.GobDecodeData(challenge.Signature, &res.Signature)
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

	if !challenge.ValidateCollect(context) {
		return
	}
	if err := challengedb.SetChallengeStatus(context, challengeID, challengedb.StatusCollected); err != nil {
		api.AbortError(context, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return
	}

	response := collectResponseFromChallenge(challenge)
	key, err := userdb.GetKey(context, response.Signature.RawID)
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
		_ = token.Set("uv", response.Signature.Response.AuthenticatorData.Flags.UserVerified())
		_ = token.Set("up", response.Signature.Response.AuthenticatorData.Flags.UserPresent())
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
		if !challenge.ValidateCollect(context) {
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
		userKey, err := userdb.GetKey(context, response.Signature.RawID)
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
			sessionID = sess.ID
			refreshToken, err = sess.CreateRefreshToken(appKey)
			if err != nil {
				api.AbortError(context, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
				return
			}
		}

		if slices.Contains(scopes, "openid") {
			// TODO: add ACR data
			idToken, err = createIDToken(context, sessionID, userKey.UserID, oauth2Ctx.Get("nonce"), app, appKey, response)
			if err != nil {
				return
			}
		}

		accessToken, err = createAccessToken(context, sessionID, userKey.UserID, appKey, app, response)
		if err != nil {
			return
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
			if err != nil {
				api.AbortError(context, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
				return
			}
		}
		accessToken, err = createAccessToken(context, session.ID, session.UserID, appKey, app, nil)
		if err != nil {
			api.AbortError(context, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
			return
		}
	case discovery.GrantTypeCIBA:
		// TODO: Implement collect for CIBA flow
	}

	context.JSON(http.StatusOK, &TokenResponse{
		AccessToken:  accessToken,
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
