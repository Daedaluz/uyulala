package public

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func AddRoutes(g *gin.RouterGroup) {
	g.Handle(http.MethodGet, "/challenge/:id", getChallengeHandler)
	g.Handle(http.MethodPost, "/challenge/:id", signChallengeHandler)
	g.Handle(http.MethodDelete, "/challenge/:id", rejectChallengeHandler)
	g.Handle(http.MethodPost, "/oauth2", createOAuth2ChallengeHandler)

	g.Handle(http.MethodGet, "/remote/:id", remoteHandler)

	g.Handle(http.MethodGet, "/jwkset.json", handleJWKSetRequest)

	g.Handle(http.MethodGet, "/aaguid/:aaguid", aaguidHandler)
}
