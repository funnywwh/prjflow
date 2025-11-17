package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"project-management/internal/utils"
	"project-management/pkg/auth"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Error(c, 401, "未授权，请先登录")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.Error(c, 401, "无效的授权头")
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := auth.ParseToken(token)
		if err != nil {
			utils.Error(c, 401, "无效的Token")
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)

		c.Next()
	}
}

