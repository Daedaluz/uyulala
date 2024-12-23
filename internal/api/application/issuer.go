package application

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"uyulala/internal/api"
	"uyulala/internal/db/keydb"

	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/spf13/viper"
)

func IssuerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
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
		set, err := keydb.GetKeys(c)
		if err != nil {
			api.AbortError(c, http.StatusInternalServerError, "no_server_keys", "No server keys could be loaded", err)
			return
		}
		keys, err := set.Set()
		if err != nil {
			api.AbortError(c, http.StatusUnauthorized, "unauthorized", "Unauthorized", err)
			return
		}
		token, err := jwt.Parse([]byte(fields[1]),
			jwt.WithValidate(true),
			jwt.WithKeySet(keys),
			jwt.WithIssuer(viper.GetString("issuer")),
			jwt.WithAcceptableSkew(time.Minute))
		if err != nil {
			api.AbortError(c, http.StatusUnauthorized, "unauthorized", "Unauthorized", err)
			return
		}
		msg, err := jws.ParseString(fields[1])
		if err != nil {
			api.AbortError(c, http.StatusUnauthorized, "unauthorized", "Unauthorized", err)
			return
		}
		tokenType, ok := msg.Signatures()[0].ProtectedHeaders().Get("typ")
		if !ok {
			api.AbortError(c, http.StatusUnauthorized, "bad_token_type", "Provided token is not an access token", err)
			return
		}
		if !strings.EqualFold(strings.ToLower(tokenType.(string)), "at+jwt") {
			api.AbortError(c, http.StatusUnauthorized, "bad_token_type", "Provided token type is not at+jwt", err)
			return
		}
		c.Set("jwt", token)
		c.Next()
	}
}
