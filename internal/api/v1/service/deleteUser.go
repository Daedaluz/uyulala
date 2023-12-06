package service

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"uyulala/internal/api"
	"uyulala/internal/db/userdb"
)

type deleteUserHandlerRequest struct {
	UserID string `json:"userId"`
}

func deleteUserHandler(ctx *gin.Context) {
	var req deleteUserHandlerRequest
	if err := ctx.BindJSON(&req); err != nil {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "Invalid request", err)
		return
	}
	if err := userdb.DeleteUser(ctx, req.UserID); err != nil {
		api.AbortError(ctx, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return
	}
	api.DeletedResponse(ctx)
}
