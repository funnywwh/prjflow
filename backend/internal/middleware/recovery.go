package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"project-management/internal/utils"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Printf("Panic recovered: %v", recovered)
		utils.Error(c, utils.CodeError, "Internal server error")
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

