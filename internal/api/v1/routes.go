package v1

import (
	"github.com/gin-gonic/gin"
	"uyulala/internal/api/application"
	"uyulala/internal/api/v1/client"
	"uyulala/internal/api/v1/oidc"
	"uyulala/internal/api/v1/public"
	"uyulala/internal/api/v1/service"
	"uyulala/internal/api/v1/user"
)

func AddRoutes(g *gin.RouterGroup) {
	privateGroup := g.Group("/")
	privateGroup.Use(application.ClientMiddleware())

	serviceGroup := g.Group("/service")
	serviceGroup.Use(
		application.ClientMiddleware(),
		application.AdminMiddleware(),
	)

	userGroup := g.Group("/user")
	userGroup.Use(application.UserMiddleware())

	issuerGroup := g.Group("/oidc")
	issuerGroup.Use(application.IssuerMiddleware())

	public.AddRoutes(g)
	client.AddRoutes(privateGroup)
	service.AddRoutes(serviceGroup)
	user.AddRoutes(userGroup)
	oidc.AddRoutes(issuerGroup)
}
