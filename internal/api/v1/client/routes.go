package client

import (
	"github.com/gin-gonic/gin"
)

func AddRoutes(g *gin.RouterGroup) {
	g.POST("/sign", createChallengeHandler)
	g.POST("/collect", collectHandler)
	g.GET("/mds/:aaguid", aaguidHandler)
}
