package middleware

import (
	"time"

	"project-management/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成请求ID
		requestID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		// 开始时间
		start := time.Now()

		// 处理请求
		c.Next()

		// 计算耗时
		latency := time.Since(start)

		// 记录日志
		if utils.Logger != nil {
			utils.Logger.WithFields(map[string]interface{}{
				"request_id": requestID,
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
				"status":     c.Writer.Status(),
				"latency":    latency,
				"ip":         c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
			}).Info("HTTP Request")
		}
	}
}

