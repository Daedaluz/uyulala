package oidc

import (
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"uyulala/internal/api"
	"uyulala/internal/api/application"
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
		"email": subj[:10] + "@uyulala.local",
	})
}
