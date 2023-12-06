package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func ChallengeResponse(ctx *gin.Context, challengeID string) {
	ctx.JSON(200, gin.H{
		"challenge_id": challengeID,
	})
}

func StatusResponse(ctx *gin.Context, code int, status, msg string) {
	ctx.AbortWithStatusJSON(code, gin.H{
		"status": status,
		"msg":    msg,
	})
}

func JSONResponse(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusOK, data)
}

func RedirectResponse(ctx *gin.Context, url string) {
	ctx.JSON(http.StatusOK, gin.H{
		"redirect": url,
	})
}

func DeletedResponse(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status": "deleted",
	})
}
