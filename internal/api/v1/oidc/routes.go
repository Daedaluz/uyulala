package oidc

import (
	"github.com/gin-gonic/gin"
)

func AddRoutes(g *gin.RouterGroup) {
	g.GET("/userinfo", userinfo)
	g.OPTIONS("/userinfo", func(context *gin.Context) {})
	g.GET("/jwkset.json", handleJWKSetRequest)
	g.OPTIONS("/jwkset.json", func(context *gin.Context) {})
}
