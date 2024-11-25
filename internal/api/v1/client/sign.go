package client

import (
	"net/http"
	"slices"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode/utf8"
	"uyulala/internal/api"
	"uyulala/internal/api/application"
	"uyulala/internal/authn"
	"uyulala/internal/db/challengedb"
	"uyulala/internal/db/userdb"
	"uyulala/openid/discovery"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/spf13/viper"
)

type CreateBIDChallengeRequest struct {
	ctx              *gin.Context                         `json:"-"`
	UserID           string                               `json:"userId"`
	UserVerification protocol.UserVerificationRequirement `json:"userVerification"`
	Text             string                               `json:"text"`
	Data             []byte                               `json:"data"`
	Timeout          uint64                               `json:"timeout"`
	Redirect         string                               `json:"redirect"`
}

type CreateCIBAChallengeRequest struct {
	ctx    *gin.Context
	UserID string
}

type CIBAAuthenticationResponse struct {
	RequestID string `json:"auth_req_id"`
	ExpiresIn uint64 `json:"expires_in"`
	Interval  uint64 `json:"interval,omitempty"`
	QRData    string `json:"qr_data,omitempty"`   // CIBA Extension
	QRSecret  string `json:"qr_secret,omitempty"` // CIBA Extension
}

func (c CreateCIBAChallengeRequest) WebAuthnID() []byte {
	return []byte(c.UserID)
}

func (c CreateCIBAChallengeRequest) WebAuthnName() string {
	return ""
}

func (c CreateCIBAChallengeRequest) WebAuthnDisplayName() string {
	return ""
}

func (c CreateCIBAChallengeRequest) WebAuthnCredentials() []webauthn.Credential {
	return userdb.GetUserKeys(c.ctx, c.UserID)
}

func (c CreateCIBAChallengeRequest) WebAuthnIcon() string {
	return ""
}

func (c CreateBIDChallengeRequest) WebAuthnID() []byte {
	return []byte(c.UserID)
}

func (c CreateBIDChallengeRequest) WebAuthnName() string {
	return ""
}

func (c CreateBIDChallengeRequest) WebAuthnDisplayName() string {
	return ""
}

func (c CreateBIDChallengeRequest) WebAuthnCredentials() []webauthn.Credential {
	return userdb.GetUserKeys(c.ctx, c.UserID)
}

func (c CreateBIDChallengeRequest) WebAuthnIcon() string {
	return ""
}

func createBIDChallenge(ctx *gin.Context) {
	req := &CreateBIDChallengeRequest{
		ctx:              ctx,
		UserID:           "",
		UserVerification: "required",
		Timeout:          5 * 60,
	}
	if err := ctx.BindJSON(req); err != nil {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "Invalid request", err)
		return
	}
	app := application.GetCurrentApplication(ctx)
	if req.Redirect != "" && !api.AllowedRedirect(app, req.Redirect) {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "Redirect not allowed", nil)
		return
	}

	if len(req.Data) > 0 && req.Text == "" {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "If data is provided, text is required too", nil)
		return
	}

	if req.Text != "" && !utf8.ValidString(req.Text) {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "Invalid text, must be utf8", nil)
		return
	}

	opts := []webauthn.LoginOption{
		webauthn.WithUserVerification(req.UserVerification),
	}
	if req.UserID != "" {
		keys, err := userdb.GetUserKeyDescriptors(ctx, req.UserID)
		if err != nil {
			api.AbortError(ctx, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
			return
		}
		if len(keys) == 0 {
			api.AbortError(ctx, http.StatusBadRequest, "no_keys", "User has no keys", nil)
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
		api.AbortError(ctx, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return
	}
	challenge, secret, err := challengedb.CreateChallenge(ctx, "webauthn.get", app.ID,
		time.Now().Add(time.Duration(req.Timeout)*time.Second), login, sessionData, req.Text, req.Data, req.Redirect)
	if err != nil {
		api.AbortError(ctx, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return
	}
	api.ChallengeResponse(ctx, challenge, secret)
}

func getUserHint(ctx *gin.Context) (string, error) {
	loginHint := ctx.Request.Form.Get("login_hint")
	loginHintToken := ctx.Request.Form.Get("login_hint_token")
	idTokenHint := ctx.Request.Form.Get("id_token_hint")
	if loginHint != "" {
		return loginHint, nil
	}
	if idTokenHint != "" || loginHintToken != "" {
		var token jwt.Token
		var err error
		if idTokenHint != "" {
			token, err = jwt.Parse([]byte(idTokenHint), jwt.WithValidate(true))
		} else {
			token, err = jwt.Parse([]byte(loginHintToken), jwt.WithValidate(true))
		}
		if err != nil {
			api.AbortError(ctx, http.StatusBadRequest, "expired_login_hint_token", "Invalid token", err)
			return "", err
		}
		if sub, ok := token.Get("sub"); ok {
			return sub.(string), nil
		}
	}
	return "", nil
}

func createCIBAChallenge(ctx *gin.Context) {
	if err := ctx.Request.ParseForm(); err != nil {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "Invalid request", err)
		return
	}
	app := application.GetCurrentApplication(ctx)
	form := ctx.Request.Form
	scopes := strings.FieldsFunc(form.Get("scope"), func(r rune) bool {
		switch r {
		case ' ', '\t', '\r', '\n':
			return true
		}
		return false
	})
	if len(scopes) == 0 {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "Scope is required", nil)
		return
	}
	clientNotificationToken := form.Get("client_notification_token")
	acrValues := strings.FieldsFunc(form.Get("acr_values"), func(r rune) bool {
		switch r {
		case ' ', '\t', '\r', '\n':
			return true
		}
		return false
	})

	bindingMessage := form.Get("binding_message")
	requestedExpiry := form.Get("requested_expiry")
	if !slices.Contains(scopes, "openid") {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "Scope does not contain openid", nil)
		return
	}
	if (app.CIBAMode == "ping" || app.CIBAMode == "push") && clientNotificationToken == "" {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "Client notification token is required when client is in push or ping mode", nil)
		return
	}
	var opts []webauthn.LoginOption
	userVerification := "preferred"
	if slices.Contains(acrValues, discovery.ACRUserPresence) {
		userVerification = "discouraged"
	}
	if slices.Contains(acrValues, discovery.ACRPreferUserVerification) {
		userVerification = "preferred"
	}
	if slices.Contains(acrValues, discovery.ACRUserVerification) {
		userVerification = "required"
	}
	opts = append(opts, webauthn.WithUserVerification(protocol.UserVerificationRequirement(userVerification)))
	var loginHint string
	var err error
	if loginHint, err = getUserHint(ctx); err != nil {
		return
	} else if loginHint != "" {
		keys, err := userdb.GetUserKeyDescriptors(ctx, loginHint)
		if err != nil {
			api.AbortError(ctx, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
			return
		}
		if len(keys) == 0 {
			api.AbortError(ctx, http.StatusBadRequest, "unknown_user_id", "Couldn't find the hinted user or the user does not have any associated keys", nil)
			return
		}
		opts = append(opts, webauthn.WithAllowedCredentials(keys))
	}

	cfg := authn.CreateWebauthnConfig()
	var login *protocol.CredentialAssertion
	var sessionData *webauthn.SessionData
	if loginHint != "" {
		login, sessionData, err = cfg.BeginLogin(&CreateCIBAChallengeRequest{ctx: ctx, UserID: loginHint}, opts...)
	} else {
		login, sessionData, err = cfg.BeginDiscoverableLogin(opts...)
	}
	if err != nil {
		api.AbortError(ctx, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return
	}

	timeout := uint64(5 * 60)
	if requestedExpiry != "" {
		i, err := strconv.ParseUint(requestedExpiry, 0, 64)
		if err != nil {
			api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "Error parsing requested_expiry", err)
			return
		}
		timeout = i
	}

	if bindingMessage != "" && !utf8.ValidString(bindingMessage) {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "Invalid binding_message, must be utf8", nil)
		return
	}

	challenge, secret, err := challengedb.CreateChallenge(ctx, "webauthn.get", app.ID,
		time.Now().Add(time.Duration(timeout)*time.Second), login, sessionData, bindingMessage, nil, "")
	if err != nil {
		api.AbortError(ctx, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return
	}

	if err := challengedb.SetOAuth2Context(ctx, challenge, form.Encode()); err != nil {
		api.AbortError(ctx, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return
	}

	cibaRequestID, err := challengedb.CreateCIBARequestID(ctx, challenge)
	if err != nil {
		api.AbortError(ctx, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return
	}

	urlTemplate := template.New("")
	urlTemplate = urlTemplate.Delims("{", "}")
	urlTemplate, err = urlTemplate.Parse(viper.GetString("ciba.qrTemplate"))
	if err != nil {
		api.AbortError(ctx, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return
	}
	var url strings.Builder
	if err := urlTemplate.Execute(&url, map[string]string{"challengeId": challenge}); err != nil {
		api.AbortError(ctx, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return
	}

	resp := &CIBAAuthenticationResponse{
		RequestID: cibaRequestID,
		ExpiresIn: timeout,
		QRSecret:  secret,
		QRData:    challenge,
	}
	if app.CIBAMode == "poll" || app.CIBAMode == "ping" {
		resp.Interval = 1
	}
	api.JSONResponse(ctx, resp)
}

func createChallengeHandler(ctx *gin.Context) {
	if ctx.Request.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		createCIBAChallenge(ctx)
	} else {
		createBIDChallenge(ctx)
	}
}
