package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"project-management/internal/api"
	"project-management/internal/config"
	"project-management/internal/middleware"
	"project-management/internal/utils"
	"project-management/internal/websocket"

	"github.com/gin-gonic/gin"
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

	// 将数据库连接存储到上下文（供中间件使用）
	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

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
	r.GET("/MP_verify_:code", wechatVerifyHandler.HandleVerifyFile)       // 匹配 /MP_verify_xxxxx.txt
	r.POST("/api/wechat/verify-file", wechatVerifyHandler.SaveVerifyFile) // 保存验证文件内容

	// 系统初始化路由（不需要认证）
	initHandler := api.NewInitHandler(db)
	initCallbackHandler := api.NewInitCallbackHandler(db)
	initGroup := r.Group("/api/init")
	{
		initGroup.GET("/status", initHandler.CheckInitStatus)
		initGroup.POST("/wechat-config", initHandler.SaveWeChatConfig)  // 第一步：保存微信配置
		initGroup.GET("/qrcode", initHandler.GetInitQRCode)             // 获取初始化二维码
		initGroup.GET("/callback", initCallbackHandler.HandleCallback)  // 微信回调处理
		initGroup.POST("", initHandler.InitSystem)                      // 第二步：通过微信登录完成初始化（保留兼容）
		initGroup.POST("/password", initHandler.InitSystemWithPassword) // 第二步：通过密码登录完成初始化
	}

	// 认证相关路由
	authHandler := api.NewAuthHandler(db)
	userHandler := api.NewUserHandler(db)
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/login", authHandler.Login) // 用户名密码登录
		authGroup.GET("/wechat/qrcode", authHandler.GetQRCode)
		authGroup.GET("/wechat/callback", authHandler.WeChatCallback)                   // 微信登录回调接口（GET请求，微信直接重定向到这里）
		authGroup.GET("/wechat/add-user/callback", userHandler.AddUserByWeChatCallback) // 微信添加用户回调接口（GET请求，微信直接重定向到这里）
		authGroup.POST("/wechat/login", authHandler.WeChatLogin)                        // 微信登录（POST请求，保留用于其他场景）
		authGroup.GET("/user/info", middleware.Auth(), authHandler.GetUserInfo)
		authGroup.POST("/logout", middleware.Auth(), authHandler.Logout)
		authGroup.POST("/change-password", middleware.Auth(), authHandler.ChangePassword) // 修改密码
	}

	// 权限管理路由
	permHandler := api.NewPermissionHandler(db)
	permGroup := r.Group("/api/permissions", middleware.Auth())
	{
		// 角色管理
		permGroup.GET("/roles", permHandler.GetRoles)
		permGroup.GET("/roles/:id", permHandler.GetRole)
		permGroup.POST("/roles", middleware.RequirePermission(db, "permission:manage"), permHandler.CreateRole)
		permGroup.PUT("/roles/:id", middleware.RequirePermission(db, "permission:manage"), permHandler.UpdateRole)
		permGroup.DELETE("/roles/:id", middleware.RequirePermission(db, "permission:manage"), permHandler.DeleteRole)
		permGroup.GET("/roles/:id/permissions", permHandler.GetRolePermissions)
		permGroup.POST("/roles/:id/permissions", middleware.RequirePermission(db, "permission:manage"), permHandler.AssignRolePermissions)

		// 权限管理
		permGroup.GET("/permissions", permHandler.GetPermissions)
		permGroup.GET("/permissions/:id", permHandler.GetPermission)
		permGroup.POST("/permissions", middleware.RequirePermission(db, "permission:manage"), permHandler.CreatePermission)
		permGroup.PUT("/permissions/:id", middleware.RequirePermission(db, "permission:manage"), permHandler.UpdatePermission)
		permGroup.DELETE("/permissions/:id", middleware.RequirePermission(db, "permission:manage"), permHandler.DeletePermission)

		// 用户角色管理
		permGroup.GET("/users/:id/roles", permHandler.GetUserRoles)
		permGroup.POST("/users/:id/roles", middleware.RequirePermission(db, "permission:manage"), permHandler.AssignUserRoles)

		// 当前用户权限
		permGroup.GET("/me", permHandler.GetUserPermissions)

		// 获取菜单树
		permGroup.GET("/menus", permHandler.GetMenus)
		permGroup.GET("/menus/all", middleware.RequirePermission(db, "permission:manage"), permHandler.GetAllMenus)
	}

	// 用户管理路由
	userGroup := r.Group("/api/users", middleware.Auth())
	{
		userGroup.GET("", middleware.RequirePermission(db, "user:read"), userHandler.GetUsers)                      // 查看用户列表
		userGroup.GET("/:id", middleware.RequirePermission(db, "user:read"), userHandler.GetUser)                   // 查看用户详情
		userGroup.POST("", middleware.RequirePermission(db, "user:create"), userHandler.CreateUser)                 // 创建用户需要权限
		userGroup.PUT("/:id", middleware.RequirePermission(db, "user:update"), userHandler.UpdateUser)              // 更新用户需要权限
		userGroup.DELETE("/:id", middleware.RequirePermission(db, "user:delete"), userHandler.DeleteUser)           // 删除用户需要权限
		userGroup.POST("/wechat/add", middleware.RequirePermission(db, "user:create"), userHandler.AddUserByWeChat) // 扫码添加用户需要权限
	}

	// 部门管理路由
	deptHandler := api.NewDepartmentHandler(db)
	deptGroup := r.Group("/api/departments", middleware.Auth())
	{
		deptGroup.GET("", middleware.RequirePermission(db, "department:read"), deptHandler.GetDepartments)
		deptGroup.POST("", middleware.RequirePermission(db, "department:create"), deptHandler.CreateDepartment)
		// 部门成员管理（必须在 /:id 之前，因为路由按顺序匹配）
		deptGroup.GET("/:id/members", middleware.RequirePermission(db, "department:read"), deptHandler.GetDepartmentMembers)
		deptGroup.POST("/:id/members", middleware.RequirePermission(db, "department:update"), deptHandler.AddDepartmentMembers)
		deptGroup.DELETE("/:id/members/:user_id", middleware.RequirePermission(db, "department:update"), deptHandler.RemoveDepartmentMember)
		// 部门CRUD（放在成员管理之后）
		deptGroup.GET("/:id", middleware.RequirePermission(db, "department:read"), deptHandler.GetDepartment)
		deptGroup.PUT("/:id", middleware.RequirePermission(db, "department:update"), deptHandler.UpdateDepartment)
		deptGroup.DELETE("/:id", middleware.RequirePermission(db, "department:delete"), deptHandler.DeleteDepartment)
	}

	// 个人工作台路由
	dashboardHandler := api.NewDashboardHandler(db)
	dashboardGroup := r.Group("/api/dashboard", middleware.Auth())
	{
		dashboardGroup.GET("", dashboardHandler.GetDashboard)
	}

	// 标签管理路由（标签是系统资源，使用项目权限）
	tagHandler := api.NewTagHandler(db)
	tagGroup := r.Group("/api/tags", middleware.Auth())
	{
		tagGroup.GET("", middleware.RequirePermission(db, "project:read"), tagHandler.GetTags)
		tagGroup.GET("/:id", middleware.RequirePermission(db, "project:read"), tagHandler.GetTag)
		tagGroup.POST("", middleware.RequirePermission(db, "project:update"), tagHandler.CreateTag)
		tagGroup.PUT("/:id", middleware.RequirePermission(db, "project:update"), tagHandler.UpdateTag)
		tagGroup.DELETE("/:id", middleware.RequirePermission(db, "project:delete"), tagHandler.DeleteTag)
	}

	// 项目管理路由
	projectHandler := api.NewProjectHandler(db)

	// 看板管理路由（需要在项目路由之前定义，因为项目路由中会用到）
	boardHandler := api.NewBoardHandler(db)

	projectGroup := r.Group("/api/projects", middleware.Auth())
	{
		projectGroup.GET("", middleware.RequirePermission(db, "project:read"), projectHandler.GetProjects)
		// 注意：统计接口、看板接口和甘特图接口需要在详情接口之前，避免路由冲突
		projectGroup.GET("/:id/statistics", middleware.RequirePermission(db, "project:read"), projectHandler.GetProjectStatistics)
		projectGroup.GET("/:id/progress", middleware.RequirePermission(db, "project:read"), projectHandler.GetProjectProgress)
		projectGroup.GET("/:id/gantt", middleware.RequirePermission(db, "project:read"), projectHandler.GetProjectGantt)
		// 项目看板路由（需要在详情路由之前）
		projectGroup.GET("/:id/boards", middleware.RequirePermission(db, "project:read"), boardHandler.GetProjectBoards)
		projectGroup.POST("/:id/boards", middleware.RequirePermission(db, "project:manage"), boardHandler.CreateBoard)
		projectGroup.GET("/:id", middleware.RequirePermission(db, "project:read"), projectHandler.GetProject)
		projectGroup.POST("", middleware.RequirePermission(db, "project:create"), projectHandler.CreateProject)
		projectGroup.PUT("/:id", middleware.RequirePermission(db, "project:update"), projectHandler.UpdateProject)
		projectGroup.DELETE("/:id", middleware.RequirePermission(db, "project:delete"), projectHandler.DeleteProject)
		// 项目成员管理
		projectGroup.GET("/:id/members", middleware.RequirePermission(db, "project:read"), projectHandler.GetProjectMembers)
		projectGroup.POST("/:id/members", middleware.RequirePermission(db, "project:manage"), projectHandler.AddProjectMembers)
		projectGroup.PUT("/:id/members/:member_id", middleware.RequirePermission(db, "project:manage"), projectHandler.UpdateProjectMember)
		projectGroup.DELETE("/:id/members/:member_id", middleware.RequirePermission(db, "project:manage"), projectHandler.RemoveProjectMember)
	}

	// 需求管理路由
	requirementHandler := api.NewRequirementHandler(db)
	requirementGroup := r.Group("/api/requirements", middleware.Auth())
	{
		requirementGroup.GET("/statistics", middleware.RequirePermission(db, "requirement:read"), requirementHandler.GetRequirementStatistics)
		requirementGroup.GET("", middleware.RequirePermission(db, "requirement:read"), requirementHandler.GetRequirements)
		requirementGroup.GET("/:id", middleware.RequirePermission(db, "requirement:read"), requirementHandler.GetRequirement)
		requirementGroup.POST("", middleware.RequirePermission(db, "requirement:create"), requirementHandler.CreateRequirement)
		requirementGroup.PUT("/:id", middleware.RequirePermission(db, "requirement:update"), requirementHandler.UpdateRequirement)
		requirementGroup.DELETE("/:id", middleware.RequirePermission(db, "requirement:delete"), requirementHandler.DeleteRequirement)
		requirementGroup.PATCH("/:id/status", middleware.RequirePermission(db, "requirement:update"), requirementHandler.UpdateRequirementStatus)
	}

	// 功能模块管理路由（模块是系统资源，使用项目权限）
	moduleHandler := api.NewModuleHandler(db)
	moduleGroup := r.Group("/api/modules", middleware.Auth())
	{
		moduleGroup.GET("", middleware.RequirePermission(db, "project:read"), moduleHandler.GetModules)
		moduleGroup.GET("/:id", middleware.RequirePermission(db, "project:read"), moduleHandler.GetModule)
		moduleGroup.POST("", middleware.RequirePermission(db, "project:update"), moduleHandler.CreateModule)
		moduleGroup.PUT("/:id", middleware.RequirePermission(db, "project:update"), moduleHandler.UpdateModule)
		moduleGroup.DELETE("/:id", middleware.RequirePermission(db, "project:delete"), moduleHandler.DeleteModule)
	}

	// Bug管理路由
	bugHandler := api.NewBugHandler(db)
	bugGroup := r.Group("/api/bugs", middleware.Auth())
	{
		bugGroup.GET("/statistics", middleware.RequirePermission(db, "bug:read"), bugHandler.GetBugStatistics)
		bugGroup.GET("", middleware.RequirePermission(db, "bug:read"), bugHandler.GetBugs)
		bugGroup.GET("/:id", middleware.RequirePermission(db, "bug:read"), bugHandler.GetBug)
		bugGroup.POST("", middleware.RequirePermission(db, "bug:create"), bugHandler.CreateBug)
		bugGroup.PUT("/:id", middleware.RequirePermission(db, "bug:update"), bugHandler.UpdateBug)
		bugGroup.DELETE("/:id", middleware.RequirePermission(db, "bug:delete"), bugHandler.DeleteBug)
		bugGroup.PATCH("/:id/status", middleware.RequirePermission(db, "bug:update"), bugHandler.UpdateBugStatus)
		bugGroup.POST("/:id/assign", middleware.RequirePermission(db, "bug:assign"), bugHandler.AssignBug)
	}

	// 任务管理路由
	taskHandler := api.NewTaskHandler(db)
	taskGroup := r.Group("/api/tasks", middleware.Auth())
	{
		taskGroup.GET("", middleware.RequirePermission(db, "task:read"), taskHandler.GetTasks)
		taskGroup.GET("/:id", middleware.RequirePermission(db, "task:read"), taskHandler.GetTask)
		taskGroup.POST("", middleware.RequirePermission(db, "task:create"), taskHandler.CreateTask)
		taskGroup.PUT("/:id", middleware.RequirePermission(db, "task:update"), taskHandler.UpdateTask)
		taskGroup.DELETE("/:id", middleware.RequirePermission(db, "task:delete"), taskHandler.DeleteTask)
		taskGroup.PATCH("/:id/status", middleware.RequirePermission(db, "task:update"), taskHandler.UpdateTaskStatus)
		taskGroup.PATCH("/:id/progress", middleware.RequirePermission(db, "task:update"), taskHandler.UpdateTaskProgress)
	}

	// 看板管理路由（看板属于项目的一部分）
	boardGroup := r.Group("/api/boards", middleware.Auth())
	{
		boardGroup.GET("/:id", middleware.RequirePermission(db, "project:read"), boardHandler.GetBoard)
		boardGroup.PUT("/:id", middleware.RequirePermission(db, "project:manage"), boardHandler.UpdateBoard)
		boardGroup.DELETE("/:id", middleware.RequirePermission(db, "project:manage"), boardHandler.DeleteBoard)
		boardGroup.GET("/:id/tasks", middleware.RequirePermission(db, "project:read"), boardHandler.GetBoardTasks)
		boardGroup.PATCH("/:id/tasks/:task_id/move", middleware.RequirePermission(db, "project:manage"), boardHandler.MoveTask)
		boardGroup.POST("/:id/columns", middleware.RequirePermission(db, "project:manage"), boardHandler.CreateBoardColumn)
		boardGroup.PUT("/:id/columns/:column_id", middleware.RequirePermission(db, "project:manage"), boardHandler.UpdateBoardColumn)
		boardGroup.DELETE("/:id/columns/:column_id", middleware.RequirePermission(db, "project:manage"), boardHandler.DeleteBoardColumn)
	}

	// 版本管理路由（版本属于项目的一部分）
	versionHandler := api.NewVersionHandler(db)
	versionGroup := r.Group("/api/versions", middleware.Auth())
	{
		versionGroup.GET("", middleware.RequirePermission(db, "project:read"), versionHandler.GetVersions)
		versionGroup.GET("/:id", middleware.RequirePermission(db, "project:read"), versionHandler.GetVersion)
		versionGroup.POST("", middleware.RequirePermission(db, "project:update"), versionHandler.CreateVersion)
		versionGroup.PUT("/:id", middleware.RequirePermission(db, "project:update"), versionHandler.UpdateVersion)
		versionGroup.DELETE("/:id", middleware.RequirePermission(db, "project:delete"), versionHandler.DeleteVersion)
		versionGroup.PATCH("/:id/status", middleware.RequirePermission(db, "project:update"), versionHandler.UpdateVersionStatus)
		versionGroup.POST("/:id/release", middleware.RequirePermission(db, "project:update"), versionHandler.ReleaseVersion)
	}

	// 测试单管理路由（测试用例属于项目的一部分）
	testCaseHandler := api.NewTestCaseHandler(db)
	testCaseGroup := r.Group("/api/test-cases", middleware.Auth())
	{
		testCaseGroup.GET("/statistics", middleware.RequirePermission(db, "project:read"), testCaseHandler.GetTestCaseStatistics)
		testCaseGroup.GET("", middleware.RequirePermission(db, "project:read"), testCaseHandler.GetTestCases)
		testCaseGroup.GET("/:id", middleware.RequirePermission(db, "project:read"), testCaseHandler.GetTestCase)
		testCaseGroup.POST("", middleware.RequirePermission(db, "project:update"), testCaseHandler.CreateTestCase)
		testCaseGroup.PUT("/:id", middleware.RequirePermission(db, "project:update"), testCaseHandler.UpdateTestCase)
		testCaseGroup.DELETE("/:id", middleware.RequirePermission(db, "project:delete"), testCaseHandler.DeleteTestCase)
		testCaseGroup.PATCH("/:id/status", middleware.RequirePermission(db, "project:update"), testCaseHandler.UpdateTestCaseStatus)
	}

	// 资源管理路由 (统计、冲突检测、利用率分析)
	resourceHandler := api.NewResourceHandler(db)
	resourceGroup := r.Group("/api/resources", middleware.Auth())
	{
		resourceGroup.GET("/statistics", middleware.RequirePermission(db, "resource:read"), resourceHandler.GetResourceStatistics)
		resourceGroup.GET("/utilization", middleware.RequirePermission(db, "resource:read"), resourceHandler.GetResourceUtilization)
		resourceGroup.GET("/conflict", middleware.RequirePermission(db, "resource:read"), resourceHandler.CheckResourceConflict)
		// resourceGroup.GET("", resourceHandler.GetResources) // Removed
		// resourceGroup.GET("/:id", resourceHandler.GetResource) // Removed
		// resourceGroup.POST("", resourceHandler.CreateResource) // Removed
		// resourceGroup.PUT("/:id", resourceHandler.UpdateResource) // Removed
		// resourceGroup.DELETE("/:id", resourceHandler.DeleteResource) // Removed
	}

	// 资源分配管理路由
	resourceAllocationHandler := api.NewResourceAllocationHandler(db)
	resourceAllocationGroup := r.Group("/api/resource-allocations", middleware.Auth())
	{
		resourceAllocationGroup.GET("/calendar", middleware.RequirePermission(db, "resource:read"), resourceAllocationHandler.GetResourceCalendar)
		resourceAllocationGroup.GET("/conflict", middleware.RequirePermission(db, "resource:read"), resourceAllocationHandler.CheckResourceConflict)
		resourceAllocationGroup.GET("", middleware.RequirePermission(db, "resource:read"), resourceAllocationHandler.GetResourceAllocations)
		resourceAllocationGroup.GET("/:id", middleware.RequirePermission(db, "resource:read"), resourceAllocationHandler.GetResourceAllocation)
		resourceAllocationGroup.POST("", middleware.RequirePermission(db, "resource:manage"), resourceAllocationHandler.CreateResourceAllocation)
		resourceAllocationGroup.PUT("/:id", middleware.RequirePermission(db, "resource:manage"), resourceAllocationHandler.UpdateResourceAllocation)
		resourceAllocationGroup.DELETE("/:id", middleware.RequirePermission(db, "resource:manage"), resourceAllocationHandler.DeleteResourceAllocation)
	}

	// 工作报告路由（日报和周报）
	reportHandler := api.NewReportHandler(db)
	reportGroup := r.Group("/api/reports", middleware.Auth())
	{
		reportGroup.GET("/work-summary", reportHandler.GetWorkSummary) // 获取工作内容汇总
	}
	dailyReportGroup := r.Group("/api/daily-reports", middleware.Auth())
	{
		dailyReportGroup.GET("", reportHandler.GetDailyReports)
		dailyReportGroup.GET("/:id", reportHandler.GetDailyReport)
		dailyReportGroup.POST("", reportHandler.CreateDailyReport)
		dailyReportGroup.PUT("/:id", reportHandler.UpdateDailyReport)
		dailyReportGroup.DELETE("/:id", reportHandler.DeleteDailyReport)
		dailyReportGroup.PATCH("/:id/status", reportHandler.UpdateDailyReportStatus)
		dailyReportGroup.POST("/:id/approve", reportHandler.ApproveDailyReport)
	}
	weeklyReportGroup := r.Group("/api/weekly-reports", middleware.Auth())
	{
		weeklyReportGroup.GET("", reportHandler.GetWeeklyReports)
		weeklyReportGroup.GET("/:id", reportHandler.GetWeeklyReport)
		weeklyReportGroup.POST("", reportHandler.CreateWeeklyReport)
		weeklyReportGroup.PUT("/:id", reportHandler.UpdateWeeklyReport)
		weeklyReportGroup.DELETE("/:id", reportHandler.DeleteWeeklyReport)
		weeklyReportGroup.PATCH("/:id/status", reportHandler.UpdateWeeklyReportStatus)
		weeklyReportGroup.POST("/:id/approve", reportHandler.ApproveWeeklyReport)
	}

	// 附件管理路由
	attachmentHandler := api.NewAttachmentHandler(db)
	attachmentGroup := r.Group("/api/attachments", middleware.Auth())
	{
		attachmentGroup.POST("/upload", middleware.RequirePermission(db, "attachment:upload"), attachmentHandler.UploadFile)
		attachmentGroup.GET("/:id", attachmentHandler.GetAttachment)
		attachmentGroup.GET("/:id/download", attachmentHandler.DownloadFile)
		attachmentGroup.DELETE("/:id", middleware.RequirePermission(db, "attachment:delete"), attachmentHandler.DeleteAttachment)
		attachmentGroup.GET("", attachmentHandler.GetAttachments)
		attachmentGroup.POST("/:id/attach", attachmentHandler.AttachToEntity)
	}

	// 静态文件服务（上传的文件）
	// 配置上传文件的静态服务，路径为 /uploads/*
	storagePath := config.AppConfig.Upload.StoragePath
	if !filepath.IsAbs(storagePath) {
		storagePath = filepath.Join(".", storagePath)
	}
	// 确保目录存在
	if err := os.MkdirAll(storagePath, 0755); err == nil {
		r.Static("/uploads", storagePath)
		log.Printf("上传文件静态服务目录: %s", storagePath)
	} else {
		log.Printf("警告: 无法创建上传文件目录 %s: %v", storagePath, err)
	}

	// 静态文件服务（前端构建后的文件）
	// 注意：必须在所有 API 路由之后，但在 catch-all 路由之前
	// 获取前端 dist 目录路径（支持从项目根目录或 backend 目录运行）
	var frontendDistPath string
	// 尝试从项目根目录查找
	if _, err := os.Stat("./frontend/dist"); err == nil {
		frontendDistPath = "./frontend/dist"
	} else if _, err := os.Stat("../frontend/dist"); err == nil {
		// 从 backend 目录运行时
		frontendDistPath = "../frontend/dist"
	} else {
		// 尝试获取可执行文件所在目录
		exePath, err := os.Executable()
		if err == nil {
			exeDir := filepath.Dir(exePath)
			// 尝试从可执行文件目录向上查找
			for _, path := range []string{
				filepath.Join(exeDir, "frontend", "dist"),
				filepath.Join(exeDir, "..", "frontend", "dist"),
				filepath.Join(exeDir, "..", "..", "frontend", "dist"),
			} {
				if _, err := os.Stat(path); err == nil {
					frontendDistPath = path
					break
				}
			}
		}
	}

	// 如果找到了前端 dist 目录，配置静态文件服务
	if frontendDistPath != "" {
		log.Printf("前端静态文件目录: %s", frontendDistPath)
		// 提供静态资源文件（如 JS、CSS、图片等）
		r.Static("/assets", filepath.Join(frontendDistPath, "assets"))
		r.StaticFile("/vite.svg", filepath.Join(frontendDistPath, "vite.svg"))

		// SPA 路由处理：所有非 API 路由都返回 index.html
		// 这样前端路由（如 /dashboard, /login 等）可以正常工作
		r.NoRoute(func(c *gin.Context) {
			// 如果请求的是 API 路径，返回 404
			path := c.Request.URL.Path
			if len(path) >= 4 && path[:4] == "/api" {
				c.JSON(http.StatusNotFound, gin.H{
					"code":    404,
					"message": "API endpoint not found",
				})
				return
			}
			// 否则返回前端 index.html（用于 SPA 路由）
			c.File(filepath.Join(frontendDistPath, "index.html"))
		})
	} else {
		log.Println("警告: 未找到前端 dist 目录，静态文件服务未启用")
		// 即使没有前端文件，也要处理 API 404
		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			if len(path) >= 4 && path[:4] == "/api" {
				c.JSON(http.StatusNotFound, gin.H{
					"code":    404,
					"message": "API endpoint not found",
				})
			} else {
				c.JSON(http.StatusNotFound, gin.H{
					"code":    404,
					"message": "Not found",
				})
			}
		})
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
