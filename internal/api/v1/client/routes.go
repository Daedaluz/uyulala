package client

import (
	"github.com/gin-gonic/gin"
)

func AddRoutes(g *gin.RouterGroup) {
	g.POST("/sign", createChallengeHandler)
	g.POST("/collect", collectHandler)
	g.OPTIONS("/collect", func(context *gin.Context) {})
	g.GET("/mds/:aaguid", aaguidHandler)
	g.OPTIONS("/mds/:aaguid", func(context *gin.Context) {})
}
