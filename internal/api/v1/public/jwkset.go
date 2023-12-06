package public

import (
	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/jwk"
	"log/slog"
	"net/http"
	"uyulala/internal/db/keydb"
)

func handleJWKSetRequest(c *gin.Context) {
	c.Header("Content-Type", "application/jwk-set+json")
	keys, err := keydb.GetKeys(c)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		slog.Warn("Failed to get keys", "err", err)
		return
	}
	set := jwk.NewSet()
	for _, key := range keys {
		k, err := key.GetPublicJWK()
		if err != nil {
			slog.Warn("Failed to get public JWK", "key", key.ID, "err", err)
			continue
		}
		set.Add(k)
	}
	c.JSON(http.StatusOK, set)
}
