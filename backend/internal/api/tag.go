package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"project-management/internal/model"
	"project-management/internal/utils"
)

type TagHandler struct {
	db *gorm.DB
}

func NewTagHandler(db *gorm.DB) *TagHandler {
	return &TagHandler{db: db}
}

// GetTags 获取所有标签列表
func (h *TagHandler) GetTags(c *gin.Context) {
	var tags []model.Tag
	if err := h.db.Order("name ASC").Find(&tags).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询失败")
		return
	}

	utils.Success(c, tags)
}

// GetTag 获取标签详情
func (h *TagHandler) GetTag(c *gin.Context) {
	id := c.Param("id")
	var tag model.Tag
	if err := h.db.First(&tag, id).Error; err != nil {
		utils.Error(c, 404, "标签不存在")
		return
	}

	utils.Success(c, tag)
}

// CreateTag 创建标签
func (h *TagHandler) CreateTag(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Color       string `json:"color"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	// 检查标签名称是否已存在
	var existingTag model.Tag
	if err := h.db.Where("name = ?", req.Name).First(&existingTag).Error; err == nil {
		utils.Error(c, 400, "标签名称已存在")
		return
	}

	tag := model.Tag{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
	}
	if tag.Color == "" {
		tag.Color = "blue"
	}

	if err := h.db.Create(&tag).Error; err != nil {
		utils.Error(c, utils.CodeError, "创建失败")
		return
	}

	utils.Success(c, tag)
}

// UpdateTag 更新标签
func (h *TagHandler) UpdateTag(c *gin.Context) {
	id := c.Param("id")
	var tag model.Tag
	if err := h.db.First(&tag, id).Error; err != nil {
		utils.Error(c, 404, "标签不存在")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	// 如果更新名称，检查是否与其他标签重名
	if req.Name != "" && req.Name != tag.Name {
		var existingTag model.Tag
		if err := h.db.Where("name = ? AND id != ?", req.Name, id).First(&existingTag).Error; err == nil {
			utils.Error(c, 400, "标签名称已存在")
			return
		}
		tag.Name = req.Name
	}

	if req.Description != "" {
		tag.Description = req.Description
	}
	if req.Color != "" {
		tag.Color = req.Color
	}

	if err := h.db.Save(&tag).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	utils.Success(c, tag)
}

// DeleteTag 删除标签
func (h *TagHandler) DeleteTag(c *gin.Context) {
	id := c.Param("id")
	var tag model.Tag
	if err := h.db.First(&tag, id).Error; err != nil {
		utils.Error(c, 404, "标签不存在")
		return
	}

	// 软删除
	if err := h.db.Delete(&tag).Error; err != nil {
		utils.Error(c, utils.CodeError, "删除失败")
		return
	}

	utils.Success(c, nil)
}

