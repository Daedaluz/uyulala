package oidc

import (
	"github.com/gin-gonic/gin"
	"uyulala/internal/api"
	"uyulala/internal/api/application"
)

func userinfo(c *gin.Context) {
	token := application.GetCurrentJWT(c)
	subj := token.Subject()
	api.JSONResponse(c, gin.H{
		"sub":   subj,
		"name":  subj[:8],
		"email": subj[:10] + "@uyulala.local",
	})
}
