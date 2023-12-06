package trust

import (
	"github.com/lestrrat-go/jwx/jwk"
	"golang.org/x/net/context"
	"net/http"
	"time"
	"uyulala/openid/discovery"
)

var (
	trust        string
	autoRefresh  *jwk.AutoRefresh
	openIDConfig *discovery.Full
	client       = &http.Client{
		Timeout: 5 * time.Second,
	}
)

func Configure(issuer string) error {
	var err error
	trust = issuer
	openIDConfig, err = discovery.FetchIssuer(issuer)
	if err != nil {
		return err
	}
	autoRefresh = jwk.NewAutoRefresh(context.Background())
	autoRefresh.Configure(openIDConfig.JWKSURI, jwk.WithHTTPClient(client), jwk.WithRefreshInterval(1*time.Hour))
	_, err = autoRefresh.Fetch(context.Background(), openIDConfig.JWKSURI)
	if err != nil {
		return err
	}
	return nil
}

func GetJWKSet() (jwk.Set, error) {
	var err error
	if openIDConfig == nil {
		openIDConfig, err = discovery.FetchIssuer(trust)
		if err != nil {
			return nil, err
		}
	}
	if autoRefresh == nil {
		autoRefresh = jwk.NewAutoRefresh(context.Background())
		autoRefresh.Configure(openIDConfig.JWKSURI, jwk.WithHTTPClient(client), jwk.WithRefreshInterval(1*time.Hour))
	}
	return autoRefresh.Fetch(context.Background(), openIDConfig.JWKSURI)
}
