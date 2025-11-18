package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"project-management/internal/model"
	"project-management/internal/utils"
)

type BoardHandler struct {
	db *gorm.DB
}

func NewBoardHandler(db *gorm.DB) *BoardHandler {
	return &BoardHandler{db: db}
}

// GetProjectBoards 获取项目的看板列表
func (h *BoardHandler) GetProjectBoards(c *gin.Context) {
	projectID := c.Param("id")
	var boards []model.Board
	if err := h.db.Where("project_id = ?", projectID).Preload("Columns").Order("created_at DESC").Find(&boards).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询失败")
		return
	}

	utils.Success(c, boards)
}

// GetBoard 获取看板详情（包含列和任务）
func (h *BoardHandler) GetBoard(c *gin.Context) {
	id := c.Param("id")
	var board model.Board
	if err := h.db.Preload("Project").Preload("Columns", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort ASC")
	}).First(&board, id).Error; err != nil {
		utils.Error(c, 404, "看板不存在")
		return
	}

	// 获取每个列对应的任务
	for i := range board.Columns {
		column := &board.Columns[i]
		var tasks []model.Task
		query := h.db.Where("project_id = ?", board.ProjectID)
		if column.Status != "" {
			query = query.Where("status = ?", column.Status)
		}
		query.Preload("Creator").Preload("Assignee").Order("created_at DESC").Find(&tasks)
		// 将任务添加到列中（通过JSON序列化时包含）
		// 注意：这里我们不能直接修改Column结构，需要通过响应结构来组织数据
	}

	utils.Success(c, board)
}

// CreateBoard 创建看板
func (h *BoardHandler) CreateBoard(c *gin.Context) {
	projectID := c.Param("id")

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Columns     []struct {
			Name   string `json:"name" binding:"required"`
			Color  string `json:"color"`
			Status string `json:"status" binding:"required"`
			Sort   int    `json:"sort"`
		} `json:"columns"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 验证项目是否存在
	var project model.Project
	if err := h.db.First(&project, projectID).Error; err != nil {
		utils.Error(c, 400, "项目不存在")
		return
	}

	board := model.Board{
		Name:        req.Name,
		Description: req.Description,
		ProjectID:   project.ID,
	}

	if err := h.db.Create(&board).Error; err != nil {
		utils.Error(c, utils.CodeError, "创建失败")
		return
	}

	// 创建列
	if len(req.Columns) > 0 {
		for i, col := range req.Columns {
			column := model.BoardColumn{
				Name:   col.Name,
				Color:  col.Color,
				Status: col.Status,
				Sort:   col.Sort,
				BoardID: board.ID,
			}
			if column.Sort == 0 {
				column.Sort = i
			}
			if err := h.db.Create(&column).Error; err != nil {
				utils.Error(c, utils.CodeError, "创建列失败")
				return
			}
		}
	}

	// 重新加载关联数据
	h.db.Preload("Project").Preload("Columns", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort ASC")
	}).First(&board, board.ID)

	utils.Success(c, board)
}

// UpdateBoard 更新看板
func (h *BoardHandler) UpdateBoard(c *gin.Context) {
	id := c.Param("id")
	var board model.Board
	if err := h.db.First(&board, id).Error; err != nil {
		utils.Error(c, 404, "看板不存在")
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	if req.Name != nil {
		board.Name = *req.Name
	}
	if req.Description != nil {
		board.Description = *req.Description
	}

	if err := h.db.Save(&board).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	// 重新加载关联数据
	h.db.Preload("Project").Preload("Columns", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort ASC")
	}).First(&board, board.ID)

	utils.Success(c, board)
}

// DeleteBoard 删除看板
func (h *BoardHandler) DeleteBoard(c *gin.Context) {
	id := c.Param("id")

	if err := h.db.Delete(&model.Board{}, id).Error; err != nil {
		utils.Error(c, utils.CodeError, "删除失败")
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// CreateBoardColumn 创建看板列
func (h *BoardHandler) CreateBoardColumn(c *gin.Context) {
	boardID := c.Param("id")
	var board model.Board
	if err := h.db.First(&board, boardID).Error; err != nil {
		utils.Error(c, 404, "看板不存在")
		return
	}

	var req struct {
		Name   string `json:"name" binding:"required"`
		Color  string `json:"color"`
		Status string `json:"status" binding:"required"`
		Sort   int    `json:"sort"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 如果未指定排序，自动设置为最大值+1
	if req.Sort == 0 {
		var maxSort int
		h.db.Model(&model.BoardColumn{}).Where("board_id = ?", boardID).Select("COALESCE(MAX(sort), -1)").Scan(&maxSort)
		req.Sort = maxSort + 1
	}

	column := model.BoardColumn{
		Name:   req.Name,
		Color:  req.Color,
		Status: req.Status,
		Sort:   req.Sort,
		BoardID: board.ID,
	}

	if err := h.db.Create(&column).Error; err != nil {
		utils.Error(c, utils.CodeError, "创建失败")
		return
	}

	utils.Success(c, column)
}

// UpdateBoardColumn 更新看板列
func (h *BoardHandler) UpdateBoardColumn(c *gin.Context) {
	columnID := c.Param("column_id")
	var column model.BoardColumn
	if err := h.db.First(&column, columnID).Error; err != nil {
		utils.Error(c, 404, "列不存在")
		return
	}

	var req struct {
		Name   *string `json:"name"`
		Color  *string `json:"color"`
		Status *string `json:"status"`
		Sort   *int    `json:"sort"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	if req.Name != nil {
		column.Name = *req.Name
	}
	if req.Color != nil {
		column.Color = *req.Color
	}
	if req.Status != nil {
		column.Status = *req.Status
	}
	if req.Sort != nil {
		column.Sort = *req.Sort
	}

	if err := h.db.Save(&column).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	utils.Success(c, column)
}

// DeleteBoardColumn 删除看板列
func (h *BoardHandler) DeleteBoardColumn(c *gin.Context) {
	columnID := c.Param("column_id")

	if err := h.db.Delete(&model.BoardColumn{}, columnID).Error; err != nil {
		utils.Error(c, utils.CodeError, "删除失败")
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// MoveTask 移动任务到不同列（拖拽排序）
func (h *BoardHandler) MoveTask(c *gin.Context) {
	boardID := c.Param("id")
	taskID := c.Param("task_id")

	var req struct {
		ColumnID string `json:"column_id" binding:"required"`
		Position int    `json:"position"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 获取看板
	var board model.Board
	if err := h.db.First(&board, boardID).Error; err != nil {
		utils.Error(c, 404, "看板不存在")
		return
	}

	// 获取列
	var column model.BoardColumn
	if err := h.db.Where("board_id = ? AND id = ?", boardID, req.ColumnID).First(&column).Error; err != nil {
		utils.Error(c, 404, "列不存在")
		return
	}

	// 获取任务
	var task model.Task
	if err := h.db.First(&task, taskID).Error; err != nil {
		utils.Error(c, 404, "任务不存在")
		return
	}

	// 验证任务是否属于该看板的项目
	if task.ProjectID != board.ProjectID {
		utils.Error(c, 400, "任务不属于该看板的项目")
		return
	}

	// 更新任务状态（根据列的状态）
	if column.Status != "" {
		task.Status = column.Status
		// 如果状态为done，自动设置进度为100
		if column.Status == "done" {
			task.Progress = 100
		}
	}

	if err := h.db.Save(&task).Error; err != nil {
		utils.Error(c, utils.CodeError, "移动失败")
		return
	}

	// 重新加载任务数据
	h.db.Preload("Project").Preload("Creator").Preload("Assignee").Preload("Dependencies").First(&task, task.ID)

	utils.Success(c, task)
}

// GetBoardTasks 获取看板的任务（按列分组）
func (h *BoardHandler) GetBoardTasks(c *gin.Context) {
	boardID := c.Param("id")
	var board model.Board
	if err := h.db.Preload("Columns", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort ASC")
	}).First(&board, boardID).Error; err != nil {
		utils.Error(c, 404, "看板不存在")
		return
	}

	// 获取项目的所有任务
	var allTasks []model.Task
	h.db.Where("project_id = ?", board.ProjectID).
		Preload("Creator").Preload("Assignee").
		Order("created_at DESC").
		Find(&allTasks)

	// 按列分组任务
	result := make(map[string]interface{})
	result["board"] = board
	result["tasks_by_column"] = make(map[uint][]model.Task)

	tasksByColumn := result["tasks_by_column"].(map[uint][]model.Task)
	for _, column := range board.Columns {
		tasksByColumn[column.ID] = []model.Task{}
		for _, task := range allTasks {
			if column.Status != "" && task.Status == column.Status {
				tasksByColumn[column.ID] = append(tasksByColumn[column.ID], task)
			}
		}
	}

	utils.Success(c, result)
}

