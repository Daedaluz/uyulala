package service

import (
	"log/slog"
	"net/http"
	"time"
	"uyulala/internal/api"
	"uyulala/internal/api/application"
	"uyulala/internal/authn"
	"uyulala/internal/db/challengedb"
	"uyulala/internal/db/userdb"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"
)

type CreateUserRequest struct {
	ctx *gin.Context `json:"-"`
	// SuggestedName is the suggested username for the user.
	SuggestedName string `json:"suggestedName"`
	ID            []byte `json:"-"`
	Timeout       uint64 `json:"timeout"`
	Redirect      string `json:"redirect"`
}

func (c *CreateUserRequest) WebAuthnID() []byte {
	if c.ID != nil {
		return c.ID
	}
	userID, err := userdb.CreateUser(c.ctx)
	if err != nil {
		slog.Error("CreateUserRequest.WebAuthnID", "error", err)
		api.AbortError(c.ctx, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return nil
	}
	c.ID = []byte(userID)
	return c.ID
}

func (c *CreateUserRequest) WebAuthnName() string {
	return c.SuggestedName
}

func (c *CreateUserRequest) WebAuthnDisplayName() string {
	return c.SuggestedName
}

func (c *CreateUserRequest) WebAuthnCredentials() []webauthn.Credential {
	return make([]webauthn.Credential, 0)
}

func (c *CreateUserRequest) WebAuthnIcon() string {
	return ""
}

func createUserHandler(ctx *gin.Context) {
	userRegistration := &CreateUserRequest{ctx: ctx}
	if err := ctx.BindJSON(userRegistration); err != nil {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "Invalid request", err)
		return
	}

	if userRegistration.SuggestedName == "" {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "Missing suggested_name", nil)
		return
	}

	if userRegistration.Redirect != "" {
		if !api.AllowedRedirect(application.GetCurrentApplication(ctx), userRegistration.Redirect) {
			api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "Redirect not allowed", nil)
			return
		}
	}

	cfg := authn.CreateWebauthnConfig()
	credential, sessionData, err := cfg.BeginRegistration(userRegistration)
	if ctx.IsAborted() {
		return
	}
	if err != nil {
		api.AbortError(ctx, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return
	}
	expires := time.Now().Add(time.Duration(userRegistration.Timeout) * time.Second)
	app := application.GetCurrentApplication(ctx)
	challengeID, secret, err := challengedb.CreateChallenge(ctx, "webauthn.create", app.ID, expires, credential, sessionData, "", []byte{}, userRegistration.Redirect)
	if err != nil {
		api.AbortError(ctx, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return
	}
	api.ChallengeResponse(ctx, challengeID, secret)
}
