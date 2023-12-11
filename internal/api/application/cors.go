package application

import "github.com/gin-gonic/gin"

func SameOriginPolicy() gin.HandlerFunc {
	return func(context *gin.Context) {
		// Check that the origin is the same
	}
}
