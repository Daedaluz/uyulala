package public

import (
	"encoding/base64"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"
	"uyulala/internal/api"
	"uyulala/internal/authn"
	"uyulala/internal/db/appdb"
	"uyulala/internal/db/challengedb"
	"uyulala/internal/db/userdb"
)

func parseRedirectURI(vars url.Values) (*url.URL, error) {
	uri := vars.Get("redirect_uri")
	if uri == "" {
		return nil, errors.New("missing redirect_uri")
	}
	parsed, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	if parsed.Fragment != "" {
		return nil, errors.New("invalid redirect_uri (Must not contain fragment)")
	}
	return parsed, nil
}

func createOAuth2ChallengeHandler(ctx *gin.Context) {
	if err := ctx.Request.ParseForm(); err != nil {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "Invalid request", err)
		return
	}
	form := ctx.Request.PostForm
	responseTypes := strings.FieldsFunc(form.Get("response_type"), func(r rune) bool { return r == ' ' || r == ',' || r == ';' || r == '\t' })
	if !(slices.Contains(responseTypes, "code")) {
		slog.Info("Unknown response type", "response_type", form.Get("response_type"))
		api.AbortError(ctx, http.StatusBadRequest, "bad_response_type", "Unknown response type (only \"code\" supported)", nil)
		return
	}
	clientID := form.Get("client_id")
	if clientID == "" {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "Missing client_id", nil)
		return
	}
	redirectURI, err := parseRedirectURI(form)
	if err != nil {
		if errors.Is(err, &url.Error{}) {
			api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "Invalid redirect_uri", err)
		} else {
			api.AbortError(ctx, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		}
		return
	}
	client, err := appdb.GetApplication(ctx, clientID)
	if err != nil {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_client", "Invalid client_id", err)
		return
	}

	if !api.AllowedRedirect(client, redirectURI.String()) {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "Redirect not allowed", nil)
		return
	}

	pkceMethod := form.Get("code_challenge_method")
	if pkceMethod != "" && pkceMethod != "S256" && pkceMethod != "plain" {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "Invalid code_challenge_method; \"S256\" or \"plain\" is supported", nil)
		return
	}

	if pkceMethod != "" && form.Get("code_challenge") == "" {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "code_challenge_method given, but no code_challenge", nil)
		return
	}
	if pkceMethod == "" && form.Get("code_challenge") != "" {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "code_challenge given, but no code_challenge_method", nil)
		return
	}

	if form.Get("state") == "" {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "Missing state", nil)
		return
	}

	userVerification := form.Get("user_verification")
	if userVerification == "" {
		userVerification = "preferred"
	}
	if prompt := form.Get("prompt"); prompt != "" {
		switch prompt {
		case "consent", "login":
			userVerification = "required"
		case "none":
			userVerification = "discouraged"
		}
	}
	opts := []webauthn.LoginOption{
		webauthn.WithUserVerification(protocol.UserVerificationRequirement(userVerification)),
	}
	userID := form.Get("user_id")
	if userID != "" {
		keys, err := userdb.GetUserKeyDescriptors(ctx, userID)
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

	signatureText := form.Get("text")
	signatureDataText := form.Get("data")
	var signatureData []byte
	if signatureDataText != "" {
		signatureData, err = base64.StdEncoding.DecodeString(signatureDataText)
		if err != nil {
			api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "Bad signature data encoding", err)
			return
		}
	}

	cfg := authn.CreateWebauthnConfig()

	var login *protocol.CredentialAssertion
	var session *webauthn.SessionData
	if userID != "" {
		user := &SignUser{userHandle: []byte(userID), ctx: ctx}
		login, session, err = cfg.BeginLogin(user, opts...)
	} else {
		login, session, err = cfg.BeginDiscoverableLogin(opts...)
	}
	if err != nil {
		api.AbortError(ctx, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return
	}

	challenge, err := challengedb.CreateChallenge(ctx, "webauthn.get", client.ID, time.Now().Add(time.Minute*5),
		login, session,
		signatureText, signatureData, redirectURI.String())
	if err != nil {
		api.AbortError(ctx, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return
	}

	if err := challengedb.SetOAuth2Context(ctx, challenge, form.Encode()); err != nil {
		api.AbortError(ctx, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return
	}

	api.ChallengeResponse(ctx, challenge)
}
