package client

import (
	"net/http"
	"uyulala/internal/api"
	"uyulala/internal/mds"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func aaguidHandler(ctx *gin.Context) {
	aaguid := ctx.Param("aaguid")
	id, err := uuid.Parse(aaguid)
	if err != nil {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_aaguid", "Invalid aaguid", err)
		return
	}
	meta, err := mds.Get(id)
	if err != nil {
		api.AbortError(ctx, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return
	}
	if meta == nil {
		api.AbortError(ctx, http.StatusNotFound, "no_aaguid", "No metadata for this aaguid", nil)
		return
	}
	api.JSONResponse(ctx, meta)
}
