package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"project-management/internal/model"
	"project-management/internal/utils"
)

type ModuleHandler struct {
	db *gorm.DB
}

func NewModuleHandler(db *gorm.DB) *ModuleHandler {
	return &ModuleHandler{db: db}
}

// GetModules 获取功能模块列表（系统资源，所有项目共享）
func (h *ModuleHandler) GetModules(c *gin.Context) {
	var modules []model.Module
	query := h.db.Model(&model.Module{})

	// 搜索
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 状态筛选
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// 排序
	query = query.Order("sort ASC, created_at DESC")

	if err := query.Find(&modules).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询失败")
		return
	}

	utils.Success(c, modules)
}

// GetModule 获取功能模块详情
func (h *ModuleHandler) GetModule(c *gin.Context) {
	id := c.Param("id")
	var module model.Module
	if err := h.db.First(&module, id).Error; err != nil {
		utils.Error(c, 404, "功能模块不存在")
		return
	}

	utils.Success(c, module)
}

// CreateModule 创建功能模块（系统资源）
func (h *ModuleHandler) CreateModule(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Code        string `json:"code"`
		Description string `json:"description"`
		Status      int    `json:"status"`
		Sort        int    `json:"sort"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 设置默认值
	if req.Status == 0 {
		req.Status = 1
	}

	module := model.Module{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Status:      req.Status,
		Sort:        req.Sort,
	}

	if err := h.db.Create(&module).Error; err != nil {
		// 检查是否是唯一索引冲突
		if utils.IsUniqueConstraintError(err) {
			if utils.IsUniqueConstraintOnField(err, "name") {
				utils.Error(c, 400, "模块名称已存在")
				return
			}
			if utils.IsUniqueConstraintOnField(err, "code") && req.Code != "" {
				utils.Error(c, 400, "模块编码已存在")
				return
			}
			utils.Error(c, 400, "模块名称或编码已存在")
			return
		}
		utils.Error(c, utils.CodeError, "创建失败: "+err.Error())
		return
	}

	utils.Success(c, module)
}

// UpdateModule 更新功能模块
func (h *ModuleHandler) UpdateModule(c *gin.Context) {
	id := c.Param("id")
	var module model.Module
	if err := h.db.First(&module, id).Error; err != nil {
		utils.Error(c, 404, "功能模块不存在")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Code        string `json:"code"`
		Description string `json:"description"`
		Status      *int   `json:"status"`
		Sort        *int   `json:"sort"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 更新字段
	if req.Name != "" {
		module.Name = req.Name
	}
	if req.Code != "" {
		module.Code = req.Code
	}
	if req.Description != "" {
		module.Description = req.Description
	}
	if req.Status != nil {
		module.Status = *req.Status
	}
	if req.Sort != nil {
		module.Sort = *req.Sort
	}

	if err := h.db.Save(&module).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	utils.Success(c, module)
}

// DeleteModule 删除功能模块
func (h *ModuleHandler) DeleteModule(c *gin.Context) {
	id := c.Param("id")
	var module model.Module
	if err := h.db.First(&module, id).Error; err != nil {
		utils.Error(c, 404, "功能模块不存在")
		return
	}

	// 检查是否有关联的Bug
	var bugCount int64
	h.db.Model(&model.Bug{}).Where("module_id = ?", id).Count(&bugCount)
	if bugCount > 0 {
		utils.Error(c, 400, "该功能模块下存在Bug，无法删除")
		return
	}

	if err := h.db.Delete(&module).Error; err != nil {
		utils.Error(c, utils.CodeError, "删除失败")
		return
	}

	utils.Success(c, nil)
}

