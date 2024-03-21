package application

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/spf13/viper"
	"log/slog"
	"net/http"
	"strings"
	"time"
	"uyulala/internal/api"
	"uyulala/internal/trust"
)

func UserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if viper.GetString("userApi.trustedIssuer") == "" {
			api.AbortError(c, http.StatusUnauthorized, "user_api_disabled", "User api disabled", nil)
			return
		}
		authorization := c.GetHeader("Authorization")
		fields := strings.Fields(authorization)
		if len(fields) != 2 {
			api.AbortError(c, http.StatusUnauthorized, "", "unauthorized", nil)
			return
		}
		if !strings.EqualFold(fields[0], "bearer") {
			api.AbortError(c, http.StatusUnauthorized, "unsupported_auth_method",
				fmt.Sprintf("%s authorization is not supported", fields[0]), nil)
			return
		}
		set, err := trust.GetJWKSet()
		if err != nil {
			api.AbortError(c, http.StatusUnauthorized, "unauthorized", "unauthorized", err)
			return
		}
		token, err := jwt.Parse([]byte(fields[1]),
			jwt.WithValidate(true),
			jwt.WithKeySet(set),
			jwt.WithIssuer(viper.GetString("userApi.trustedIssuer")),
			jwt.WithAcceptableSkew(time.Minute))
		if err != nil {
			api.AbortError(c, http.StatusUnauthorized, "unauthorized", "unauthorized", err)
			return
		}
		c.Set("jwt", token)
		c.Next()
	}
}

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		fields := strings.Fields(authorization)
		if len(fields) != 2 {
			slog.Warn("JWTMiddleware", "fields", len(fields))
			c.Next()
			return
		}
		if !strings.EqualFold(fields[0], "bearer") {
			slog.Warn("JWTMiddleware", "auth_type", fields[0])
			c.Next()
			return
		}
		set, err := trust.GetJWKSet()
		if err != nil {
			slog.Warn("JWTMiddleware", "jwkset_error", err)
			c.Next()
			return
		}
		token, err := jwt.Parse([]byte(fields[1]),
			jwt.WithValidate(true),
			jwt.WithKeySet(set),
			jwt.WithIssuer(viper.GetString("userApi.trustedIssuer")),
			jwt.WithAcceptableSkew(time.Minute))
		if err != nil {
			slog.Warn("JWTMiddleware", "parse_error", err)
			c.Next()
			return
		}
		c.Set("jwt", token)
	}
}

func GetCurrentJWT(c *gin.Context) jwt.Token {
	token, exists := c.Get("jwt")
	if !exists {
		return nil
	}
	return token.(jwt.Token)
}
