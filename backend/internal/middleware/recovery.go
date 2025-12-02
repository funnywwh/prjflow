package middleware

import (
	"net/http"
	"runtime/debug"

	"project-management/internal/utils"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if utils.Logger != nil {
			utils.Logger.WithFields(map[string]interface{}{
				"panic":   recovered,
				"stack":   string(debug.Stack()),
				"path":    c.Request.URL.Path,
				"method":  c.Request.Method,
			}).Error("Panic recovered")
		}
		utils.Error(c, utils.CodeError, "Internal server error")
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

