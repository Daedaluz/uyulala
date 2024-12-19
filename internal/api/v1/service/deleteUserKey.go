package service

import (
	"net/http"
	"uyulala/internal/api"
	"uyulala/internal/db/userdb"

	"github.com/gin-gonic/gin"
)

type deleteUserKeyHandlerRequest struct {
	UserID string `json:"userId"`
	KeyID  string `json:"keyHash"`
}

func deleteUserKeyHandler(ctx *gin.Context) {
	var req deleteUserKeyHandlerRequest
	if err := ctx.BindJSON(&req); err != nil {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "Invalid request", err)
		return
	}
	if err := userdb.DeleteUserKey(ctx, req.UserID, req.KeyID); err != nil {
		api.AbortError(ctx, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return
	}
	api.DeletedResponse(ctx)
}
