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
func GetPageSize(c *gin.Context) int {
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", strconv.Itoa(DefaultPageSize)))
	if err != nil || pageSize < 1 {
		return DefaultPageSize
	}
	if pageSize > MaxPageSize {
		return MaxPageSize
	}
	return pageSize
}

