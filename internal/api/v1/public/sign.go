package public

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"net/url"
	"uyulala/internal/api"
	"uyulala/internal/authn"
	"uyulala/internal/db/challengedb"
	"uyulala/internal/db/userdb"
)

type SignUser struct {
	ctx        *gin.Context
	rawID      []byte
	userHandle []byte
}

func (s *SignUser) WebAuthnID() []byte {
	return s.userHandle
}

func (s *SignUser) WebAuthnName() string {
	return ""
}

func (s *SignUser) WebAuthnDisplayName() string {
	return ""
}

func (s *SignUser) WebAuthnCredentials() []webauthn.Credential {
	return userdb.GetUserKeys(s.ctx, string(s.userHandle))
}

func (s *SignUser) WebAuthnIcon() string {
	return ""
}

func signLogin(context *gin.Context, challengeID string, challenge *challengedb.Data) {
	cfg := authn.CreateWebauthnConfig()
	session := webauthn.SessionData{}
	if err := challenge.Expand(nil, &session); err != nil {
		slog.Error("signLogin Expand session", "error", err)
		api.AbortError(context, http.StatusInternalServerError, "invalid_challenge", "Unexpected error", err)
		return
	}
	response := &protocol.CredentialAssertionResponse{}
	if err := context.BindJSON(response); err != nil {
		slog.Error("signLogin BindJSON", "error", err)
		api.AbortError(context, http.StatusBadRequest, "invalid_response", "Invalid response", err)
		return
	}
	parsed, err := response.Parse()
	if err != nil {
		slog.Error("signLogin Parse response", "error", err)
		api.AbortError(context, http.StatusBadRequest, "invalid_response", "Invalid response", err)
		return
	}
	user := &SignUser{
		userHandle: session.UserID,
		ctx:        context,
	}
	var cred *webauthn.Credential
	if session.UserID == nil {
		cred, err = cfg.ValidateDiscoverableLogin(func(rawID, userHandle []byte) (webauthn.User, error) {
			user.rawID = rawID
			user.userHandle = userHandle
			user.ctx = context
			return user, nil
		}, session, parsed)
	} else {
		cred, err = cfg.ValidateLogin(user, session, parsed)
	}
	if err != nil {
		slog.Error("signLogin ValidateLogin", "error", err)
		api.AbortError(context, http.StatusBadRequest, "invalid_response", "Invalid response", err)
		return
	}
	if context.IsAborted() {
		return
	}

	if err := userdb.PingUserKey(context, cred); err != nil {
		slog.Error("signLogin PingUserKey", "error", err)
		api.AbortError(context, http.StatusBadRequest, "invalid_response", "Invalid response", err)
		return
	}

	if err := challengedb.SignChallenge(context, challengeID, parsed, cred); err != nil {
		slog.Error("signLogin SignChallenge", "error", err)
		api.AbortError(context, http.StatusInternalServerError, "invalid_challenge", "Unexpected error", err)
		return
	}

	if cred.Flags.UserVerified {
		if err := userdb.UpdateAuthTime(context, string(user.userHandle), challenge.AppID); err != nil {
			slog.Error("signLogin UpdateAuthTime", "error", err)
			api.AbortError(context, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
			return
		}
	}

	redirectURL := ""
	if challenge.RedirectURL != "" {
		redirectURL = challenge.RedirectURL
		if r, err := url.Parse(challenge.RedirectURL); err == nil {
			q := r.Query()
			if oauthContext, err := url.ParseQuery(challenge.OAuth2Context); err == nil && len(oauthContext) > 0 {
				// TODO: check CIBA ping / push modes
				code, err := challengedb.CreateCode(context, challengeID)
				if err != nil {
					slog.Error("signLogin CreateCode", "error", err)
					api.AbortError(context, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
					return
				}
				q.Set("code", code)
				if oauthContext.Get("state") != "" {
					q.Set("state", oauthContext.Get("state"))
				}
			} else {
				q.Set("challengeId", challengeID)
			}
			r.RawQuery = q.Encode()
			redirectURL = r.String()
		}
	}
	api.RedirectResponse(context, redirectURL)
}
func signCreate(context *gin.Context, challengeID string, challenge *challengedb.Data) {
	cfg := authn.CreateWebauthnConfig()
	session := webauthn.SessionData{}
	if err := challenge.Expand(nil, &session); err != nil {
		slog.Error("signCreate Expand session", "error", err)
		api.AbortError(context, http.StatusInternalServerError, "invalid_challenge", "Unexpected error", err)
		return
	}

	response := &protocol.CredentialCreationResponse{}
	if err := context.BindJSON(response); err != nil {
		slog.Error("signCreate BindJSON", "error", err)
		api.AbortError(context, http.StatusBadRequest, "invalid_response", "Invalid response", err)
		return
	}

	parsed, err := response.Parse()
	if err != nil {
		slog.Error("signCreate Parse response", "error", err)
		api.AbortError(context, http.StatusBadRequest, "invalid_response", "Invalid response", err)
		return
	}

	usr := &SignUser{
		userHandle: session.UserID,
	}
	cred, err := cfg.CreateCredential(usr, session, parsed)
	if err != nil {
		slog.Error("signCreate CreateCredential", "error", err)
		api.AbortError(context, http.StatusBadRequest, "invalid_response", "Invalid response", err)
		return
	}

	aaguid, err := uuid.FromBytes(cred.Authenticator.AAGUID)
	if err != nil {
		slog.Error("signCreate uuid.FromBytes", "error", err)
		api.AbortError(context, http.StatusBadRequest, "invalid_response", "Invalid response", err)
		return
	}

	if err := userdb.CreateUserKey(context, string(session.UserID), aaguid, cred); err != nil {
		slog.Error("signCreate CreateUserKey", "error", err)
		api.AbortError(context, http.StatusInternalServerError, "invalid_challenge", "Unexpected error", err)
		return
	}
	if err := challengedb.DeleteChallenge(context, challengeID); err != nil {
		slog.Error("signCreate DeleteChallenge", "error", err)
		api.AbortError(context, http.StatusInternalServerError, "invalid_challenge", "Unexpected error", err)
		return
	}
	api.RedirectResponse(context, challenge.RedirectURL)
}

func signChallengeHandler(context *gin.Context) {
	challengeID := context.Param("id")
	challenge, err := challengedb.GetChallenge(context, challengeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			api.AbortError(context, http.StatusNotFound, "not_found", "Challenge not found", err)
			return
		}
		slog.Error("signChallengeHandler GetChallenge", "error", err)
		api.AbortError(context, http.StatusInternalServerError, "invalid_challenge", "Unexpected error", err)
		return
	}
	if !challenge.Validate(context) {
		return
	}

	switch challenge.Type {
	case "webauthn.get":
		signLogin(context, challengeID, challenge)
	case "webauthn.create":
		signCreate(context, challengeID, challenge)
	}
}
