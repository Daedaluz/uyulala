package service

import (
	"github.com/gin-gonic/gin"
)

func AddRoutes(g *gin.RouterGroup) {
	g.GET("/list/users", listUsersHandler)

	g.POST("/create/user", createUserHandler)
	g.POST("/create/key", createKeyHandler)

	g.POST("/delete/user", deleteUserHandler)
	g.POST("/delete/key", deleteUserKeyHandler)
}
