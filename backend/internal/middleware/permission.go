package middleware

import (
	"project-management/internal/utils"
	"project-management/pkg/permission"

	"github.com/gin-gonic/gin"
)

// RequirePermission 要求特定权限
func RequirePermission(permCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("roles")
		if !exists {
			utils.Error(c, 403, "没有权限")
			c.Abort()
			return
		}

		roleList, ok := roles.([]string)
		if !ok {
			utils.Error(c, 403, "没有权限")
			c.Abort()
			return
		}

		// 检查用户是否有权限
		hasPermission := permission.CheckPermission(roleList, permCode)
		if !hasPermission {
			utils.Error(c, 403, "没有权限")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole 要求特定角色
func RequireRole(roleCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("roles")
		if !exists {
			utils.Error(c, 403, "没有权限")
			c.Abort()
			return
		}

		roleList, ok := roles.([]string)
		if !ok {
			utils.Error(c, 403, "没有权限")
			c.Abort()
			return
		}

		// 检查用户是否有角色
		hasRole := false
		for _, role := range roleList {
			if role == roleCode {
				hasRole = true
				break
			}
		}

		if !hasRole {
			utils.Error(c, 403, "没有权限")
			c.Abort()
			return
		}

		c.Next()
	}
}

