package wellknown

import (
	"fmt"
	"log/slog"
	"net/http"
	"uyulala/internal/db/keydb"
	"uyulala/openid/discovery"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func OpenIDConfigurationHandler(c *gin.Context) {
	proto := "http"
	if c.Request.Header.Get("X-Forwarded-Proto") == "https" || c.Request.TLS != nil {
		proto = "https"
	}
	issuer := fmt.Sprintf("%s://%s", proto, c.Request.Host)
	if viper.GetString("issuer") != "" {
		issuer = viper.GetString("issuer")
	}
	req := &discovery.Required{
		Issuer:                                 issuer,
		AuthorizationEndpoint:                  fmt.Sprintf("%s/authorize", issuer),
		TokenEndpoint:                          fmt.Sprintf("%s/api/v1/collect", issuer),
		JWKSURI:                                fmt.Sprintf("%s/api/v1/oidc/jwkset.json", issuer),
		ResponseTypesSupported:                 []string{discovery.ResponseTypeCode},
		GrantTypesSupported:                    []string{discovery.GrantTypeAuthorizationCode, discovery.GrantTypeCIBA},
		ScopesSupported:                        []string{"openid", "offline_access"},
		BackChannelAuthenticationEndpoint:      fmt.Sprintf("%s/api/v1/sign", issuer),
		BackChannelTokenDeliveryModesSupported: []string{"poll", "ping", "push"},
		BackChannelAuthenticationQREndpoint:    fmt.Sprintf("%s/authenticator", issuer),
	}
	opt := &discovery.Optional{
		ACRValuesSupported: []string{
			discovery.ACRUserPresence,
			discovery.ACRPreferUserVerification,
			discovery.ACRUserVerification,
		},
		CodeChallengeMethodsSupported: []string{"plain", "S256"},
	}
	cfg := discovery.NewConfig(req, opt)
	userinfoEndpoint := viper.GetString("userInfo.endpoint")
	if userinfoEndpoint == "" {
		userinfoEndpoint = fmt.Sprintf("%s/api/v1/oidc/userinfo", issuer)
	}
	cfg.UserInfoEndpoint = userinfoEndpoint
	cfg.ResponseModesSupported = []string{discovery.ResponseModeQuery}
	cfg.TokenEndpointAuthMethodsSupported = []string{discovery.TokenAuthClientSecretPost, discovery.TokenAuthClientSecretBasic}
	algs, err := keydb.GetAvailableAlgorithms(c)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		slog.Warn("Failed to get available algorithms", "err", err)
		return
	}
	cfg.IDTokenSigningAlgValuesSupported = algs
	c.JSON(http.StatusOK, cfg)
}
