package client

import (
	"net/http"
	"uyulala/internal/api"

	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/jwt"
)

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
