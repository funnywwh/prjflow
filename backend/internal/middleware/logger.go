package middleware

import (
	"log"
	"time"

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
		log.Printf(
			"[%s] %s %s %d %v %s %s",
			requestID,
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			latency,
			c.ClientIP(),
			c.Request.UserAgent(),
		)
	}
}

