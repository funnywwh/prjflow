package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"project-management/internal/api"
	"project-management/internal/config"
	"project-management/internal/middleware"
	"project-management/internal/utils"
)

func main() {
	// 加载配置
	if err := config.LoadConfig(""); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 设置Gin模式
	gin.SetMode(config.AppConfig.Server.Mode)

	// 初始化数据库
	db, err := utils.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 自动迁移数据库
	if err := utils.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migrated successfully")

	// 创建Gin引擎
	r := gin.New()

	// 注册中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// 健康检查接口
	r.GET("/health", func(c *gin.Context) {
		utils.Success(c, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// 认证相关路由
	authHandler := api.NewAuthHandler(db)
	authGroup := r.Group("/auth")
	{
		authGroup.GET("/wechat/qrcode", authHandler.GetQRCode)
		authGroup.POST("/wechat/login", authHandler.WeChatLogin)
		authGroup.GET("/user/info", middleware.Auth(), authHandler.GetUserInfo)
		authGroup.POST("/logout", middleware.Auth(), authHandler.Logout)
	}

	// 权限管理路由
	permHandler := api.NewPermissionHandler(db)
	permGroup := r.Group("/permissions", middleware.Auth())
	{
		permGroup.GET("/roles", permHandler.GetRoles)
		permGroup.POST("/roles", permHandler.CreateRole)
		permGroup.PUT("/roles/:id", permHandler.UpdateRole)
		permGroup.DELETE("/roles/:id", permHandler.DeleteRole)
		permGroup.GET("/permissions", permHandler.GetPermissions)
		permGroup.POST("/permissions", permHandler.CreatePermission)
		permGroup.POST("/roles/:id/permissions", permHandler.AssignRolePermissions)
		permGroup.POST("/users/:id/roles", permHandler.AssignUserRoles)
	}

	// 用户管理路由
	userHandler := api.NewUserHandler(db)
	userGroup := r.Group("/users", middleware.Auth())
	{
		userGroup.GET("", userHandler.GetUsers)
		userGroup.GET("/:id", userHandler.GetUser)
		userGroup.POST("", userHandler.CreateUser)
		userGroup.PUT("/:id", userHandler.UpdateUser)
		userGroup.DELETE("/:id", userHandler.DeleteUser)
	}

	// 部门管理路由
	deptHandler := api.NewDepartmentHandler(db)
	deptGroup := r.Group("/departments", middleware.Auth())
	{
		deptGroup.GET("", deptHandler.GetDepartments)
		deptGroup.GET("/:id", deptHandler.GetDepartment)
		deptGroup.POST("", deptHandler.CreateDepartment)
		deptGroup.PUT("/:id", deptHandler.UpdateDepartment)
		deptGroup.DELETE("/:id", deptHandler.DeleteDepartment)
	}

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.AppConfig.Server.Port),
		Handler:      r,
		ReadTimeout:  time.Duration(config.AppConfig.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.AppConfig.Server.WriteTimeout) * time.Second,
	}

	// 启动服务器（异步）
	go func() {
		log.Printf("Server starting on port %d", config.AppConfig.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

