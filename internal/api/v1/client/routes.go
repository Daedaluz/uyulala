package client

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func AddRoutes(g *gin.RouterGroup) {
	g.Handle(http.MethodPost, "/sign", createChallengeHandler)
	g.Handle(http.MethodPost, "/collect", collectHandler)
}
