package oidc

import "github.com/gin-gonic/gin"

func AddRoutes(g *gin.RouterGroup) {
	g.GET("/userinfo", userinfo)
}
