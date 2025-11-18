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
	"project-management/internal/websocket"
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

	// WebSocket路由
	r.GET("/ws", websocket.HandleWebSocket)

	// 微信验证文件路由（不需要认证，必须放在根路径）
	// 支持格式：/MP_verify_xxxxx.txt
	// 注意：这个路由必须放在其他路由之前，但只匹配 MP_verify_ 开头的文件
	wechatVerifyHandler := api.NewWeChatVerifyHandler(db)
	// 使用参数路由，匹配 /MP_verify_:code.txt 格式
	// 注意：Gin 不支持在参数名中包含点，所以我们需要使用 :code 参数
	r.GET("/MP_verify_:code", wechatVerifyHandler.HandleVerifyFile) // 匹配 /MP_verify_xxxxx.txt
	r.POST("/api/wechat/verify-file", wechatVerifyHandler.SaveVerifyFile) // 保存验证文件内容

	// 系统初始化路由（不需要认证）
	initHandler := api.NewInitHandler(db)
	initCallbackHandler := api.NewInitCallbackHandler(db)
	initGroup := r.Group("/api/init")
	{
		initGroup.GET("/status", initHandler.CheckInitStatus)
		initGroup.POST("/wechat-config", initHandler.SaveWeChatConfig) // 第一步：保存微信配置
		initGroup.GET("/qrcode", initHandler.GetInitQRCode)            // 获取初始化二维码
		initGroup.GET("/callback", initCallbackHandler.HandleCallback) // 微信回调处理
		initGroup.POST("", initHandler.InitSystem)                     // 第二步：通过微信登录完成初始化（保留兼容）
	}

	// 认证相关路由
	authHandler := api.NewAuthHandler(db)
	userHandler := api.NewUserHandler(db)
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/login", authHandler.Login)                               // 用户名密码登录
		authGroup.GET("/wechat/qrcode", authHandler.GetQRCode)
		authGroup.GET("/wechat/callback", authHandler.WeChatCallback)              // 微信登录回调接口（GET请求，微信直接重定向到这里）
		authGroup.GET("/wechat/add-user/callback", userHandler.AddUserByWeChatCallback) // 微信添加用户回调接口（GET请求，微信直接重定向到这里）
		authGroup.POST("/wechat/login", authHandler.WeChatLogin)                    // 微信登录（POST请求，保留用于其他场景）
		authGroup.GET("/user/info", middleware.Auth(), authHandler.GetUserInfo)
		authGroup.POST("/logout", middleware.Auth(), authHandler.Logout)
		authGroup.POST("/change-password", middleware.Auth(), authHandler.ChangePassword) // 修改密码
	}

	// 权限管理路由
	permHandler := api.NewPermissionHandler(db)
	permGroup := r.Group("/api/permissions", middleware.Auth())
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
	userGroup := r.Group("/api/users", middleware.Auth())
	{
		userGroup.GET("", userHandler.GetUsers)
		userGroup.GET("/:id", userHandler.GetUser)
		userGroup.POST("", userHandler.CreateUser)
		userGroup.PUT("/:id", userHandler.UpdateUser)
		userGroup.DELETE("/:id", userHandler.DeleteUser)
		userGroup.POST("/wechat/add", userHandler.AddUserByWeChat) // 扫码添加用户
	}

	// 部门管理路由
	deptHandler := api.NewDepartmentHandler(db)
	deptGroup := r.Group("/api/departments", middleware.Auth())
	{
		deptGroup.GET("", deptHandler.GetDepartments)
		deptGroup.POST("", deptHandler.CreateDepartment)
		// 部门成员管理（必须在 /:id 之前，因为路由按顺序匹配）
		deptGroup.GET("/:id/members", deptHandler.GetDepartmentMembers)
		deptGroup.POST("/:id/members", deptHandler.AddDepartmentMembers)
		deptGroup.DELETE("/:id/members/:user_id", deptHandler.RemoveDepartmentMember)
		// 部门CRUD（放在成员管理之后）
		deptGroup.GET("/:id", deptHandler.GetDepartment)
		deptGroup.PUT("/:id", deptHandler.UpdateDepartment)
		deptGroup.DELETE("/:id", deptHandler.DeleteDepartment)
	}

	// 个人工作台路由
	dashboardHandler := api.NewDashboardHandler(db)
	dashboardGroup := r.Group("/api/dashboard", middleware.Auth())
	{
		dashboardGroup.GET("", dashboardHandler.GetDashboard)
	}

	// 产品管理路由
	productHandler := api.NewProductHandler(db)
	productLineGroup := r.Group("/api/product-lines", middleware.Auth())
	{
		productLineGroup.GET("", productHandler.GetProductLines)
		productLineGroup.GET("/:id", productHandler.GetProductLine)
		productLineGroup.POST("", productHandler.CreateProductLine)
		productLineGroup.PUT("/:id", productHandler.UpdateProductLine)
		productLineGroup.DELETE("/:id", productHandler.DeleteProductLine)
	}
	productGroup := r.Group("/api/products", middleware.Auth())
	{
		productGroup.GET("", productHandler.GetProducts)
		productGroup.GET("/:id", productHandler.GetProduct)
		productGroup.POST("", productHandler.CreateProduct)
		productGroup.PUT("/:id", productHandler.UpdateProduct)
		productGroup.DELETE("/:id", productHandler.DeleteProduct)
	}

	// 项目管理路由
	projectHandler := api.NewProjectHandler(db)
	projectGroupGroup := r.Group("/api/project-groups", middleware.Auth())
	{
		projectGroupGroup.GET("", projectHandler.GetProjectGroups)
		projectGroupGroup.GET("/:id", projectHandler.GetProjectGroup)
		projectGroupGroup.POST("", projectHandler.CreateProjectGroup)
		projectGroupGroup.PUT("/:id", projectHandler.UpdateProjectGroup)
		projectGroupGroup.DELETE("/:id", projectHandler.DeleteProjectGroup)
	}
	
	// 看板管理路由（需要在项目路由之前定义，因为项目路由中会用到）
	boardHandler := api.NewBoardHandler(db)
	
	projectGroup := r.Group("/api/projects", middleware.Auth())
	{
		projectGroup.GET("", projectHandler.GetProjects)
		// 注意：统计接口、看板接口和甘特图接口需要在详情接口之前，避免路由冲突
		projectGroup.GET("/:id/statistics", projectHandler.GetProjectStatistics)
		projectGroup.GET("/:id/progress", projectHandler.GetProjectProgress)
		projectGroup.GET("/:id/gantt", projectHandler.GetProjectGantt)
		// 项目看板路由（需要在详情路由之前）
		projectGroup.GET("/:id/boards", boardHandler.GetProjectBoards)
		projectGroup.POST("/:id/boards", boardHandler.CreateBoard)
		projectGroup.GET("/:id", projectHandler.GetProject)
		projectGroup.POST("", projectHandler.CreateProject)
		projectGroup.PUT("/:id", projectHandler.UpdateProject)
		projectGroup.DELETE("/:id", projectHandler.DeleteProject)
		// 项目成员管理
		projectGroup.GET("/:id/members", projectHandler.GetProjectMembers)
		projectGroup.POST("/:id/members", projectHandler.AddProjectMembers)
		projectGroup.PUT("/:id/members/:member_id", projectHandler.UpdateProjectMember)
		projectGroup.DELETE("/:id/members/:member_id", projectHandler.RemoveProjectMember)
	}

	// 需求管理路由
	requirementHandler := api.NewRequirementHandler(db)
	requirementGroup := r.Group("/api/requirements", middleware.Auth())
	{
		requirementGroup.GET("/statistics", requirementHandler.GetRequirementStatistics)
		requirementGroup.GET("", requirementHandler.GetRequirements)
		requirementGroup.GET("/:id", requirementHandler.GetRequirement)
		requirementGroup.POST("", requirementHandler.CreateRequirement)
		requirementGroup.PUT("/:id", requirementHandler.UpdateRequirement)
		requirementGroup.DELETE("/:id", requirementHandler.DeleteRequirement)
		requirementGroup.PATCH("/:id/status", requirementHandler.UpdateRequirementStatus)
	}

	// Bug管理路由
	bugHandler := api.NewBugHandler(db)
	bugGroup := r.Group("/api/bugs", middleware.Auth())
	{
		bugGroup.GET("/statistics", bugHandler.GetBugStatistics)
		bugGroup.GET("", bugHandler.GetBugs)
		bugGroup.GET("/:id", bugHandler.GetBug)
		bugGroup.POST("", bugHandler.CreateBug)
		bugGroup.PUT("/:id", bugHandler.UpdateBug)
		bugGroup.DELETE("/:id", bugHandler.DeleteBug)
		bugGroup.PATCH("/:id/status", bugHandler.UpdateBugStatus)
		bugGroup.POST("/:id/assign", bugHandler.AssignBug)
	}

	// 任务管理路由
	taskHandler := api.NewTaskHandler(db)
	taskGroup := r.Group("/api/tasks", middleware.Auth())
	{
		taskGroup.GET("", taskHandler.GetTasks)
		taskGroup.GET("/:id", taskHandler.GetTask)
		taskGroup.POST("", taskHandler.CreateTask)
		taskGroup.PUT("/:id", taskHandler.UpdateTask)
		taskGroup.DELETE("/:id", taskHandler.DeleteTask)
		taskGroup.PATCH("/:id/status", taskHandler.UpdateTaskStatus)
		taskGroup.PATCH("/:id/progress", taskHandler.UpdateTaskProgress)
	}

	// 看板管理路由
	boardGroup := r.Group("/api/boards", middleware.Auth())
	{
		boardGroup.GET("/:id", boardHandler.GetBoard)
		boardGroup.PUT("/:id", boardHandler.UpdateBoard)
		boardGroup.DELETE("/:id", boardHandler.DeleteBoard)
		boardGroup.GET("/:id/tasks", boardHandler.GetBoardTasks)
		boardGroup.PATCH("/:id/tasks/:task_id/move", boardHandler.MoveTask)
		boardGroup.POST("/:id/columns", boardHandler.CreateBoardColumn)
		boardGroup.PUT("/:id/columns/:column_id", boardHandler.UpdateBoardColumn)
		boardGroup.DELETE("/:id/columns/:column_id", boardHandler.DeleteBoardColumn)
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

