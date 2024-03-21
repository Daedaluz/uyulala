package wellknown

import "github.com/gin-gonic/gin"

func AddRoutes(g *gin.RouterGroup) {
	g.GET("/.well-known/openid-configuration", OpenIDConfigurationHandler)
	g.OPTIONS("/.well-known/openid-configuration", func(context *gin.Context) {
	})
}
