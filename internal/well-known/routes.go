package wellknown

import "github.com/gin-gonic/gin"

func AddRoutes(g *gin.RouterGroup) {
	g.GET("/.well-known/openid-configuration", OpenIDConfigurationHandler)
}
