package user

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
	ctx           *gin.Context `json:"-"`
	SuggestedName string       `json:"suggestedName"`
	ID            string       `json:"-"`
	Timeout       uint64       `json:"timeout"`
	Redirect      string       `json:"-"`
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

func addKey(c *gin.Context) {
	jwt := application.GetCurrentJWT(c)
	subj := jwt.Subject()
	req := &CreateKeyRequest{
		ctx:      c,
		ID:       subj,
		Timeout:  uint64((time.Minute * 5).Seconds()),
		Redirect: c.GetHeader("referer"),
	}
	if err := c.BindJSON(req); err != nil {
		api.AbortError(c, http.StatusBadRequest, "invalid_request", "Invalid request", err)
		return
	}
	if req.SuggestedName == "" {
		api.AbortError(c, http.StatusBadRequest, "invalid_request", "Must have a name", nil)
		return
	}
	cfg := authn.CreateWebauthnConfig()
	credential, sessionData, err := cfg.BeginRegistration(req)
	if c.IsAborted() {
		return
	}
	if err != nil {
		api.AbortError(c, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return
	}

	expires := time.Now().Add(time.Duration(req.Timeout) * time.Second)
	challengeID, secret, err := challengedb.CreateChallenge(c, "webauthn.create", "", expires, credential, sessionData, "", []byte{}, req.Redirect)
	if err != nil {
		api.AbortError(c, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return
	}
	api.ChallengeResponse(c, challengeID, secret)
}
