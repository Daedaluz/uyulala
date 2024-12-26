package oidc

import (
	"fmt"
	"log/slog"
	"net/http"
	"uyulala/internal/api"
	"uyulala/internal/api/application"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func userinfo(c *gin.Context) {
	token := application.GetCurrentJWT(c)
	if token == nil {
		slog.Info("no token provided")
		api.AbortError(c, http.StatusUnauthorized, "no_jwt", "No JWT provided", nil)
		return
	}
	subj := token.Subject()
	api.JSONResponse(c, gin.H{
		"sub":   subj,
		"name":  subj[:8],
		"email": fmt.Sprintf("%s@%s", subj[:10], viper.GetString("userInfo.emailSuffix")),
	})
}
