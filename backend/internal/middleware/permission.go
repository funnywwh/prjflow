package middleware

import (
	"project-management/internal/utils"
	"project-management/pkg/permission"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RequirePermission 要求特定权限
func RequirePermission(db *gorm.DB, permCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先从上下文获取权限列表（如果已加载）
		if perms, exists := c.Get("permissions"); exists {
			if permList, ok := perms.([]string); ok {
				// 检查权限列表中是否包含所需权限
				for _, perm := range permList {
					if perm == permCode {
						c.Next()
						return
					}
				}
				utils.Error(c, 403, "没有权限")
				c.Abort()
				return
			}
		}

		// 如果上下文没有权限列表，从角色查询
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

		// 如果用户没有任何角色，直接拒绝访问
		if len(roleList) == 0 {
			utils.Error(c, 403, "没有权限：用户未分配角色")
			c.Abort()
			return
		}

		// 检查用户是否有权限
		hasPermission, err := permission.CheckPermissionWithDB(db, roleList, permCode)
		if err != nil {
			utils.Error(c, utils.CodeError, "权限检查失败")
			c.Abort()
			return
		}

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

