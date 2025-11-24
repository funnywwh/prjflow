package middleware

import (
	"strings"

	"project-management/internal/model"
	"project-management/internal/utils"
	"project-management/pkg/auth"
	"project-management/pkg/permission"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

		// 从上下文获取数据库连接（如果存在）
		if db, exists := c.Get("db"); exists {
			if dbConn, ok := db.(*gorm.DB); ok {
				// 加载用户权限到上下文（提高性能）
				// 如果用户有角色，加载角色权限；如果没有角色，设置空权限列表
				if len(claims.Roles) > 0 {
					permCodes, err := permission.GetRolePermissions(dbConn, claims.Roles)
					if err == nil {
						c.Set("permissions", permCodes)
					} else {
						// 如果加载失败，设置空权限列表
						c.Set("permissions", []string{})
					}
				} else {
					// 用户没有角色，设置空权限列表
					c.Set("permissions", []string{})
				}
			}
		}

		c.Next()
	}
}

// AuthWithDB 带数据库连接的认证中间件（用于加载用户权限）
func AuthWithDB(db *gorm.DB) gin.HandlerFunc {
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

		// 加载用户权限到上下文
		// 如果用户有角色，加载角色权限；如果没有角色，设置空权限列表
		if len(claims.Roles) > 0 {
			permCodes, err := permission.GetRolePermissions(db, claims.Roles)
			if err == nil {
				c.Set("permissions", permCodes)
			} else {
				// 如果加载失败，设置空权限列表
				c.Set("permissions", []string{})
			}
		} else {
			// 用户没有角色，设置空权限列表
			c.Set("permissions", []string{})
		}

		c.Next()
	}
}

// AuthOptional 可选认证中间件：如果系统已初始化，需要认证；如果未初始化，允许访问
func AuthOptional(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查系统是否已初始化
		var initConfig model.SystemConfig
		isInitialized := db.Where("key = ?", "initialized").First(&initConfig).Error == nil && initConfig.Value == "true"

		// 如果系统未初始化，跳过认证
		if !isInitialized {
			c.Next()
			return
		}

		// 如果系统已初始化，执行认证
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

		// 加载用户权限到上下文
		if len(claims.Roles) > 0 {
			permCodes, err := permission.GetRolePermissions(db, claims.Roles)
			if err == nil {
				c.Set("permissions", permCodes)
			} else {
				c.Set("permissions", []string{})
			}
		} else {
			c.Set("permissions", []string{})
		}

		c.Next()
	}
}