package v1

import (
	"github.com/gin-gonic/contrib/cors"
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

	publicGroup := g.Group("/")
	publicGroup.Use(cors.New(cors.Config{
		AbortOnError:     true,
		AllowAllOrigins:  false,
		AllowedOrigins:   viper.GetStringSlice("webauthn.origins"),
		AllowOriginFunc:  nil,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodDelete},
		AllowedHeaders:   nil,
		ExposedHeaders:   nil,
		AllowCredentials: false,
		MaxAge:           0,
	}))

	clientGroup := g.Group("/")
	clientGroup.Use(
		application.ClientMiddleware(),
		cors.New(cors.Config{
			AllowAllOrigins: true,
			AllowedHeaders:  []string{"Authorization", "*"},
		}),
	)

	serviceGroup := g.Group("/service")
	serviceGroup.Use(
		application.ClientMiddleware(),
		application.AdminMiddleware(),
		cors.New(cors.Config{
			AllowAllOrigins: true,
			AllowedHeaders:  []string{"Authorization", "*"},
		}),
	)

	userGroup := g.Group("/user")
	userGroup.Use(
		application.UserMiddleware(),
		cors.New(cors.Config{
			AbortOnError:     true,
			AllowAllOrigins:  false,
			AllowedOrigins:   viper.GetStringSlice("webauthn.origins"),
			AllowOriginFunc:  nil,
			AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodDelete},
			AllowedHeaders:   nil,
			ExposedHeaders:   nil,
			AllowCredentials: false,
			MaxAge:           0,
		}),
	)

	issuerGroup := g.Group("/oidc")
	issuerGroup.Use(application.IssuerMiddleware())

	public.AddRoutes(publicGroup)
	client.AddRoutes(clientGroup)
	service.AddRoutes(serviceGroup)
	user.AddRoutes(userGroup)
	oidc.AddRoutes(issuerGroup)
}
