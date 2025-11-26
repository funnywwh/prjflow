package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// GetPage 获取页码
func GetPage(c *gin.Context) int {
	page, err := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(DefaultPage)))
	if err != nil || page < 1 {
		return DefaultPage
	}
	return page
}

// GetPageSize 获取每页数量
// 同时支持 page_size 和 size 参数，优先使用 page_size
func GetPageSize(c *gin.Context) int {
	// 优先使用 page_size 参数
	pageSizeStr := c.Query("page_size")
	if pageSizeStr == "" {
		// 如果没有 page_size，尝试使用 size 参数
		pageSizeStr = c.Query("size")
	}
	
	// 如果都没有，使用默认值
	if pageSizeStr == "" {
		return DefaultPageSize
	}
	
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		return DefaultPageSize
	}
	if pageSize > MaxPageSize {
		return MaxPageSize
	}
	return pageSize
}

