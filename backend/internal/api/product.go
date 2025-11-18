package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"project-management/internal/model"
	"project-management/internal/utils"
)

type ProductHandler struct {
	db *gorm.DB
}

func NewProductHandler(db *gorm.DB) *ProductHandler {
	return &ProductHandler{db: db}
}

// GetProductLines 获取产品线列表
func (h *ProductHandler) GetProductLines(c *gin.Context) {
	var productLines []model.ProductLine
	query := h.db

	// 搜索
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 状态筛选
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&productLines).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询失败")
		return
	}

	utils.Success(c, productLines)
}

// GetProductLine 获取产品线详情
func (h *ProductHandler) GetProductLine(c *gin.Context) {
	id := c.Param("id")
	var productLine model.ProductLine
	if err := h.db.Preload("Products").First(&productLine, id).Error; err != nil {
		utils.Error(c, 404, "产品线不存在")
		return
	}

	utils.Success(c, productLine)
}

// CreateProductLine 创建产品线
func (h *ProductHandler) CreateProductLine(c *gin.Context) {
	var productLine model.ProductLine
	if err := c.ShouldBindJSON(&productLine); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	if err := h.db.Create(&productLine).Error; err != nil {
		utils.Error(c, utils.CodeError, "创建失败")
		return
	}

	utils.Success(c, productLine)
}

// UpdateProductLine 更新产品线
func (h *ProductHandler) UpdateProductLine(c *gin.Context) {
	id := c.Param("id")
	var productLine model.ProductLine
	if err := h.db.First(&productLine, id).Error; err != nil {
		utils.Error(c, 404, "产品线不存在")
		return
	}

	if err := c.ShouldBindJSON(&productLine); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	if err := h.db.Save(&productLine).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	utils.Success(c, productLine)
}

// DeleteProductLine 删除产品线
func (h *ProductHandler) DeleteProductLine(c *gin.Context) {
	id := c.Param("id")

	// 检查是否有产品
	var count int64
	h.db.Model(&model.Product{}).Where("product_line_id = ?", id).Count(&count)
	if count > 0 {
		utils.Error(c, 400, "产品线下存在产品，无法删除")
		return
	}

	if err := h.db.Delete(&model.ProductLine{}, id).Error; err != nil {
		utils.Error(c, utils.CodeError, "删除失败")
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// GetProducts 获取产品列表
func (h *ProductHandler) GetProducts(c *gin.Context) {
	var products []model.Product
	query := h.db.Preload("ProductLine")

	// 搜索
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 产品线筛选
	if productLineID := c.Query("product_line_id"); productLineID != "" {
		query = query.Where("product_line_id = ?", productLineID)
	}

	// 状态筛选
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// 分页
	page := utils.GetPage(c)
	pageSize := utils.GetPageSize(c)
	offset := (page - 1) * pageSize

	var total int64
	query.Model(&model.Product{}).Count(&total)

	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&products).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询失败")
		return
	}

	utils.Success(c, gin.H{
		"list":  products,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// GetProduct 获取产品详情
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id := c.Param("id")
	var product model.Product
	if err := h.db.Preload("ProductLine").Preload("Projects").First(&product, id).Error; err != nil {
		utils.Error(c, 404, "产品不存在")
		return
	}

	utils.Success(c, product)
}

// CreateProduct 创建产品
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var product model.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 验证产品线是否存在
	if product.ProductLineID > 0 {
		var productLine model.ProductLine
		if err := h.db.First(&productLine, product.ProductLineID).Error; err != nil {
			utils.Error(c, 404, "产品线不存在")
			return
		}
	}

	if err := h.db.Create(&product).Error; err != nil {
		utils.Error(c, utils.CodeError, "创建失败")
		return
	}

	// 重新加载关联数据
	h.db.Preload("ProductLine").First(&product, product.ID)

	utils.Success(c, product)
}

// UpdateProduct 更新产品
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var product model.Product
	if err := h.db.First(&product, id).Error; err != nil {
		utils.Error(c, 404, "产品不存在")
		return
	}

	if err := c.ShouldBindJSON(&product); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 验证产品线是否存在
	if product.ProductLineID > 0 {
		var productLine model.ProductLine
		if err := h.db.First(&productLine, product.ProductLineID).Error; err != nil {
			utils.Error(c, 404, "产品线不存在")
			return
		}
	}

	if err := h.db.Save(&product).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	// 重新加载关联数据
	h.db.Preload("ProductLine").First(&product, product.ID)

	utils.Success(c, product)
}

// DeleteProduct 删除产品
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")

	// 检查是否有项目关联
	var count int64
	h.db.Model(&model.Project{}).Where("product_id = ?", id).Count(&count)
	if count > 0 {
		utils.Error(c, 400, "产品下存在项目，无法删除")
		return
	}

	if err := h.db.Delete(&model.Product{}, id).Error; err != nil {
		utils.Error(c, utils.CodeError, "删除失败")
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

