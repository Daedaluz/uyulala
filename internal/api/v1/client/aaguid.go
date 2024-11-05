package client

import (
	"net/http"
	"sync"
	"uyulala/internal/api"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/metadata"
	"github.com/google/uuid"
)

var (
	lock = sync.Mutex{}
	meta map[uuid.UUID]*metadata.Entry
)

func getMeta() (map[uuid.UUID]*metadata.Entry, error) {
	lock.Lock()
	defer lock.Unlock()
	m := meta
	if m == nil {
		tmp, err := metadata.Fetch()
		if err != nil {
			return nil, err
		}
		m = tmp.ToMap()
	}
	return m, nil
}

func init() {
	_, _ = getMeta()
}

func aaguidHandler(ctx *gin.Context) {
	aaguid := ctx.Param("aaguid")
	id, err := uuid.Parse(aaguid)
	if err != nil {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_aaguid", "Invalid aaguid", err)
		return
	}
	var metaMap map[uuid.UUID]*metadata.Entry
	if metaMap, err = getMeta(); err != nil {
		api.AbortError(ctx, http.StatusBadGateway, "fetch", "unable to get production metadata from fido alliance", err)
		return
	}

	meta, ok := metaMap[id]
	if !ok {
		api.AbortError(ctx, http.StatusNotFound, "no_aaguid", "No metadata for this aaguid", nil)
		return
	}
	api.JSONResponse(ctx, meta)
}
