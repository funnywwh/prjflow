package utils

import (
	"project-management/internal/model"

	"gorm.io/gorm"
)

// AutoMigrate 自动迁移所有模型
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		// 用户与权限
		&model.User{},
		&model.Department{},
		&model.Role{},
		&model.Permission{},

		// 产品与项目
		&model.ProductLine{},
		&model.Product{},
		&model.ProjectGroup{},
		&model.Project{},
		&model.ProjectMember{},

		// 需求与Bug
		&model.Requirement{},
		&model.Bug{},
		&model.BugAssignee{},

		// 任务与看板
		&model.Task{},
		&model.TaskDependency{},
		&model.Board{},
		&model.BoardColumn{},

		// 计划与执行
		&model.Plan{},
		&model.PlanExecution{},

		// 版本与构建
		&model.Build{},
		&model.Version{},

		// 测试
		&model.TestCase{},
		&model.TestReport{},
		&model.TestCaseBug{},
		&model.TestCaseReport{},

		// 资源管理
		&model.Resource{},
		&model.ResourceAllocation{},

		// 工作报告
		&model.DailyReport{},
		&model.WeeklyReport{},

		// 插件管理
		&model.Plugin{},
		&model.PluginConfig{},
		&model.PluginHook{},
		&model.PluginRoute{},

		// 关系图
		&model.EntityRelation{},

		// 工作台
		&model.UserDashboard{},
	)
}

