package v1

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
	"uyulala/internal/api/application"
	"uyulala/internal/api/v1/client"
	"uyulala/internal/api/v1/oidc"
	"uyulala/internal/api/v1/public"
	"uyulala/internal/api/v1/service"
	"uyulala/internal/api/v1/user"
)

func AddRoutes(g *gin.RouterGroup) {
	publicCorsConfig := cors.Config{
		AllowAllOrigins:  false,
		AllowOrigins:     viper.GetStringSlice("webauthn.origins"),
		AllowOriginFunc:  nil,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodDelete},
		AllowHeaders:     nil,
		ExposeHeaders:    nil,
		AllowCredentials: false,
		MaxAge:           0,
	}
	clientCorsConfig := cors.Config{
		AllowAllOrigins:  true,
		AllowHeaders:     []string{"Authorization", "*"},
		AllowCredentials: true,
	}

	publicGroup := g.Group("/")
	publicGroup.Use(cors.New(publicCorsConfig))

	clientGroup := g.Group("/")
	clientGroup.Use(
		application.ClientMiddleware(),
		cors.New(clientCorsConfig),
	)

	serviceGroup := g.Group("/service")
	serviceGroup.Use(
		application.ClientMiddleware(),
		application.AdminMiddleware(),
		cors.New(clientCorsConfig),
	)

	userGroup := g.Group("/user")
	userGroup.Use(
		application.UserMiddleware(),
		cors.New(clientCorsConfig),
	)

	issuerGroup := g.Group("/oidc")
	issuerGroup.Use(
		cors.New(clientCorsConfig),
	)

	public.AddRoutes(publicGroup)
	client.AddRoutes(clientGroup)
	service.AddRoutes(serviceGroup)
	user.AddRoutes(userGroup)
	oidc.AddRoutes(issuerGroup)
}
