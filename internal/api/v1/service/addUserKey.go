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

type CreateKeyRequest struct {
	ctx           *gin.Context
	SuggestedName string `json:"suggestedName"`
	ID            string `json:"userId"`
	Timeout       int64  `json:"timeout"`
	Redirect      string `json:"redirect"`
}

func (c *CreateKeyRequest) WebAuthnID() []byte {
	user, err := userdb.GetUser(c.ctx, c.ID)
	if err != nil {
		slog.Error("CreateUserRequest.WebAuthnID", "error", err)
		api.AbortError(c.ctx, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return nil
	}
	return []byte(user.ID)
}

func (c *CreateKeyRequest) WebAuthnName() string {
	return c.SuggestedName
}

func (c *CreateKeyRequest) WebAuthnDisplayName() string {
	return c.SuggestedName
}

func (c *CreateKeyRequest) WebAuthnCredentials() []webauthn.Credential {
	return make([]webauthn.Credential, 0)
}

func (c *CreateKeyRequest) WebAuthnIcon() string {
	return ""
}

func createKeyHandler(ctx *gin.Context) {
	req := &CreateKeyRequest{ctx: ctx}
	if err := ctx.BindJSON(req); err != nil {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "Invalid request", err)
		return
	}
	if req.SuggestedName == "" {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "Must have a name", nil)
		return
	}
	cfg := authn.CreateWebauthnConfig()
	credential, sessionData, err := cfg.BeginRegistration(req)
	if ctx.IsAborted() {
		return
	}
	if err != nil {
		api.AbortError(ctx, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return
	}

	if req.Redirect != "" {
		if !api.AllowedRedirect(application.GetCurrentApplication(ctx), req.Redirect) {
			api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "Redirect not allowed", nil)
			return
		}
	}

	expires := time.Now().Add(time.Duration(req.Timeout).Abs() * time.Second)
	app := application.GetCurrentApplication(ctx)
	challengeID, secret, err := challengedb.CreateChallenge(ctx, &challengedb.CreateChallengeData{
		Type:          "webauthn.create",
		AppID:         app.ID,
		Expire:        expires,
		PublicData:    credential,
		PrivateData:   sessionData,
		Nonce:         "",
		SignatureText: "",
		SignatureData: nil,
		RedirectURL:   req.Redirect,
	}, "")
	if err != nil {
		api.AbortError(ctx, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return
	}
	api.ChallengeResponse(ctx, challengeID, secret)
}
