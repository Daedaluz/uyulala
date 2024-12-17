package public

import (
	"github.com/gin-gonic/gin"
)

func AddRoutes(g *gin.RouterGroup) {
	g.POST("/challenge", getChallengeHandlerPost)
	g.PUT("/challenge", signChallengeHandler)
	g.DELETE("/challenge", rejectChallengeHandler)

	g.POST("/oauth2", createOAuth2ChallengeHandler)
}
