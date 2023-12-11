package public

import (
	"github.com/gin-gonic/gin"
)

func AddRoutes(g *gin.RouterGroup) {
	g.GET("/challenge/:id", getChallengeHandler)
	g.POST("/challenge/:id", signChallengeHandler)
	g.DELETE("/challenge/:id", rejectChallengeHandler)
	g.POST("/oauth2", createOAuth2ChallengeHandler)

	g.GET("/remote/:id", remoteHandler)
}
