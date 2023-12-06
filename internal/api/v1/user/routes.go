package user

import (
	"github.com/gin-gonic/gin"
)

func AddRoutes(g *gin.RouterGroup) {
	g.POST("/addKey", addKey)
	g.POST("/deleteKey", deleteKey)
	g.GET("/listKeys", listKeys)
}
