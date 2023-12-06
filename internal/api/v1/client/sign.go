package client

import (
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"log/slog"
	"net/http"
	"time"
	"uyulala/internal/api"
	"uyulala/internal/api/application"
	"uyulala/internal/authn"
	"uyulala/internal/db/challengedb"
	"uyulala/internal/db/userdb"
)

type CreateChallengeRequest struct {
	ctx              *gin.Context                         `json:"-"`
	UserID           string                               `json:"userId"`
	UserVerification protocol.UserVerificationRequirement `json:"userVerification"`
	Text             string                               `json:"text"`
	Data             []byte                               `json:"data"`
	Timeout          uint64                               `json:"timeout"`
	Redirect         string                               `json:"redirect"`
}

func (c CreateChallengeRequest) WebAuthnID() []byte {
	return []byte(c.UserID)
}

func (c CreateChallengeRequest) WebAuthnName() string {
	return ""
}

func (c CreateChallengeRequest) WebAuthnDisplayName() string {
	return ""
}

func (c CreateChallengeRequest) WebAuthnCredentials() []webauthn.Credential {
	return userdb.GetUserKeys(c.ctx, c.UserID)
}

func (c CreateChallengeRequest) WebAuthnIcon() string {
	return ""
}

func createChallengeHandler(context *gin.Context) {
	req := &CreateChallengeRequest{
		ctx:              context,
		UserID:           "",
		UserVerification: "required",
		Timeout:          5 * 60,
	}
	if err := context.BindJSON(req); err != nil {
		api.AbortError(context, http.StatusBadRequest, "invalid_request", "Invalid request", err)
		return
	}
	app := application.GetCurrentApplication(context)
	if req.Redirect != "" && !api.AllowedRedirect(app, req.Redirect) {
		api.AbortError(context, http.StatusBadRequest, "invalid_request", "Redirect not allowed", nil)
		return
	}

	if len(req.Data) > 0 && req.Text == "" {
		api.AbortError(context, http.StatusBadRequest, "invalid_request", "If data is provided, text is required too", nil)
		return
	}

	opts := []webauthn.LoginOption{
		webauthn.WithUserVerification(req.UserVerification),
	}
	if req.UserID != "" {
		keys, err := userdb.GetUserKeyDescriptors(context, req.UserID)
		if err != nil {
			api.AbortError(context, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
			return
		}
		if len(keys) == 0 {
			api.AbortError(context, http.StatusBadRequest, "no_keys", "User has no keys", nil)
			return
		}
		opts = append(opts, webauthn.WithAllowedCredentials(keys))
	}

	cfg := authn.CreateWebauthnConfig()

	var login *protocol.CredentialAssertion
	var sessionData *webauthn.SessionData
	var err error
	if req.UserID != "" {
		login, sessionData, err = cfg.BeginLogin(req, opts...)
	} else {
		login, sessionData, err = cfg.BeginDiscoverableLogin(opts...)
	}
	if err != nil {
		api.AbortError(context, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return
	}
	slog.Info("Text", "text", req.Text, "data", req.Data)
	challenge, err := challengedb.CreateChallenge(context, "webauthn.get", app.ID,
		time.Now().Add(time.Duration(req.Timeout)*time.Second), login, sessionData, req.Text, req.Data, req.Redirect)
	if err != nil {
		api.AbortError(context, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return
	}
	api.ChallengeResponse(context, challenge)
}
