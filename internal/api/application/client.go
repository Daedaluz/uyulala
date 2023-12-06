package application

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/spf13/viper"
	"log/slog"
	"net/http"
	"strings"
	"uyulala/internal/api"
	"uyulala/internal/db/appdb"
	"uyulala/internal/db/challengedb"
	"uyulala/internal/db/keydb"
	"uyulala/internal/db/sessiondb"
	"uyulala/internal/db/userdb"
)

func authOAuthCollect(ctx *gin.Context, app *appdb.Application, challenge *challengedb.Data, codeVerifier string) {
	oauth2Context := challenge.GetOAuth2Context()
	if method := oauth2Context.Get("code_challenge_method"); method != "" {
		codeChallenge := oauth2Context.Get("code_challenge")
		switch method {
		case "S256":
			sha256Hash := sha256.Sum256([]byte(codeVerifier))
			codeChallengeHash := base64.RawURLEncoding.EncodeToString(sha256Hash[:])
			if subtle.ConstantTimeCompare([]byte(codeChallengeHash), []byte(codeChallenge)) == 0 {
				api.AbortError(ctx, http.StatusBadRequest, "invalid_challenge", "Bad code verifier", nil)
				return
			}
		case "plain":
			if subtle.ConstantTimeCompare([]byte(codeVerifier), []byte(codeChallenge)) == 0 {
				api.AbortError(ctx, http.StatusBadRequest, "invalid_challenge", "Bad code verifier", nil)
				return
			}
		}
	}
	ctx.Set("application", app)
	ctx.Set("challenge", challenge)
}

func authCollect(ctx *gin.Context, app *appdb.Application, password string) {
	if subtle.ConstantTimeCompare([]byte(password), []byte(app.Secret)) == 0 {
		api.AbortError(ctx, http.StatusUnauthorized, "unauthorized", "Invalid credentials", nil)
		return
	}
	ctx.Set("application", app)
}

func authOAuthRefresh(ctx *gin.Context, app *appdb.Application) {
	refreshToken := ctx.PostForm("refresh_token")
	keys, err := keydb.GetKeys(ctx)
	if err != nil {
		api.AbortError(ctx, http.StatusInternalServerError, "db_error", "Unable to fetch server keys", err)
		return
	}
	keySet, err := keys.Set()
	if err != nil {
		api.AbortError(ctx, http.StatusInternalServerError, "key_error", "Unable to get server key set", err)
		return
	}
	token, err := jwt.ParseString(refreshToken, jwt.WithValidate(true),
		jwt.WithIssuer(viper.GetString("issuer")),
		jwt.WithKeySet(keySet), jwt.WithAudience(viper.GetString("issuer")))
	if err != nil {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_token", "Invalid token", err)
		return
	}

	msg, err := jws.ParseString(refreshToken)
	if err != nil {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_token", "Invalid token", err)
		return
	}
	if typ, ok := msg.Signatures()[0].ProtectedHeaders().Get("typ"); ok && strings.ToLower(typ.(string)) != "refresh+jwt" {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_token", "This isn't a refresh token", err)
		return
	}
	tid := token.JwtID()

	sess, err := sessiondb.Get(ctx, tid)
	if err != nil {
		api.AbortError(ctx, http.StatusBadRequest, "no_session", "Session not found", nil)
		return
	}

	if c, ok := token.Get("counter"); ok {
		if sess.Counter != uint32(c.(float64)) {
			slog.Warn("Cloned refresh token", "token", tid, "counter", sess.Counter, "expected", uint32(c.(float32)))
			api.AbortError(ctx, http.StatusBadRequest, "reused_refresh_token", "Reused refresh token", nil)
			_ = sessiondb.Delete(ctx, tid)
			return
		}
	} else {
		slog.Warn("Refresh token without counter", "token", tid)
		api.AbortError(ctx, http.StatusBadRequest, "malformed_refresh_token", "Refresh token lacks counter", nil)
		return
	}

	sessApp, err := appdb.GetApplication(ctx, sess.AppID)
	if err != nil {
		api.AbortError(ctx, http.StatusInternalServerError, "no_app", "Couldn't find application in session", err)
		return
	}

	if sessApp.ID != app.ID {
		api.AbortError(ctx, http.StatusBadRequest, "wrong_app", "This refresh token is for another application", nil)
		return
	}

	user, err := userdb.GetUser(ctx, sess.UserID)
	if err != nil {
		api.AbortError(ctx, http.StatusInternalServerError, "no_user", "Couldn't find user in session", err)
		return
	}

	ctx.Set("application", app)
	ctx.Set("user", user)
	ctx.Set("session", sess)
}

func ClientMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		username, password, ok := ctx.Request.BasicAuth()
		if ctx.Request.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
			if !ok {
				username = ctx.PostForm("client_id")
				password = ctx.PostForm("client_secret")
			}
			app, err := appdb.GetApplication(ctx, username)
			if err != nil {
				api.AbortError(ctx, http.StatusUnauthorized, "unauthorized", "Invalid credentials", err)
				return
			}
			if subtle.ConstantTimeCompare([]byte(password), []byte(app.Secret)) == 0 {
				api.AbortError(ctx, http.StatusUnauthorized, "unauthorized", "Invalid credentials", nil)
				return
			}
			grantType := ctx.PostForm("grant_type")
			switch grantType {
			case "authorization_code":
				code := ctx.PostForm("code")
				codeVerifier := ctx.PostForm("code_verifier")
				ch, err := challengedb.GetChallenge(ctx, code)
				if err != nil {
					api.AbortError(ctx, http.StatusUnauthorized, "no_challenge", "Invalid challenge", err)
					return
				}
				if ch.AppID != app.ID {
					api.AbortError(ctx, http.StatusUnauthorized, "no_challenge", "Challenge wasn't for this client", err)
					return
				}
				authOAuthCollect(ctx, app, ch, codeVerifier)
			case "refresh_token":
				authOAuthRefresh(ctx, app)
			default:
				api.AbortError(ctx, http.StatusBadRequest, "invalid_grant_type", fmt.Sprintf("Unsupported grant type %s", grantType), nil)
			}
		} else {
			if !ok {
				api.AbortError(ctx, http.StatusUnauthorized, "unauthorized", "Basic auth required", nil)
				return
			}
			app, err := appdb.GetApplication(ctx, username)
			if err != nil {
				api.AbortError(ctx, http.StatusUnauthorized, "unauthorized", "Invalid credentials", err)
				return
			}
			authCollect(ctx, app, password)
		}
	}
}

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		app := GetCurrentApplication(c)
		if !app.Admin {
			api.AbortError(c, http.StatusUnauthorized, "not_admin", "This endpoint requires an administrative app", nil)
			return
		}
	}
}

func GetCurrentApplication(ctx *gin.Context) *appdb.Application {
	app, _ := ctx.Get("application")
	return app.(*appdb.Application)
}

func GetCurrentChallenge(ctx *gin.Context) *challengedb.Data {
	challenge, exists := ctx.Get("challenge")
	if !exists {
		return nil
	}
	return challenge.(*challengedb.Data)
}

func GetCurrentSession(ctx *gin.Context) *sessiondb.Session {
	session, _ := ctx.Get("session")
	return session.(*sessiondb.Session)
}
