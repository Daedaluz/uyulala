package client

import (
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/metadata"
	"github.com/google/uuid"
	"net/http"
	"uyulala/internal/api"
)

func aaguidHandler(ctx *gin.Context) {
	aaguid := ctx.Param("aaguid")
	id, err := uuid.Parse(aaguid)
	if err != nil {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_aaguid", "Invalid aaguid", err)
		return
	}
	meta, ok := metadata.Metadata[id]
	if !ok {
		api.AbortError(ctx, http.StatusNotFound, "no_aaguid", "No metadata for this aaguid", nil)
		return
	}
	api.JSONResponse(ctx, meta)
}
