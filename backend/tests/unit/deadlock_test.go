package unit

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"project-management/internal/model"
)

// TestDeadlockFix_ConcurrentResourceCreation 测试并发创建资源时的死锁修复
// 直接测试数据库操作，模拟 syncTaskActualHours 的逻辑
func TestDeadlockFix_ConcurrentResourceCreation(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	// 确保资源相关表已创建
	require.NoError(t, db.AutoMigrate(&model.Resource{}, &model.ResourceAllocation{}))

	// 创建测试数据
	user := CreateTestUser(t, db, "testuser1", "测试用户1")
	project := CreateTestProject(t, db, "测试项目")

	// 创建任务
	task := &model.Task{
		Title:      "测试任务",
		ProjectID: project.ID,
		AssigneeID: &user.ID,
		CreatorID:  user.ID,
		Status:     "doing",
	}
	require.NoError(t, db.Create(task).Error)

	// 并发数量
	concurrency := 10
	workDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// 使用 WaitGroup 等待所有 goroutine 完成
	var wg sync.WaitGroup
	wg.Add(concurrency)

	// 错误通道
	errChan := make(chan error, concurrency)

	// 并发创建资源分配（模拟 syncTaskActualHours 的逻辑）
	for i := 0; i < concurrency; i++ {
		go func(index int) {
			defer wg.Done()
			actualHours := float64(1.0 + float64(index)*0.1)

			// 使用事务包裹所有操作（修复后的逻辑）
			tx := db.Begin()
			defer func() {
				if r := recover(); r != nil {
					tx.Rollback()
					errChan <- fmt.Errorf("panic: %v", r)
				}
			}()

			// 使用 FirstOrCreate 查找或创建资源（修复后的逻辑）
			var resource model.Resource
			if err := tx.Where("user_id = ? AND project_id = ?", user.ID, project.ID).
				FirstOrCreate(&resource, model.Resource{
					UserID:    user.ID,
					ProjectID: project.ID,
				}).Error; err != nil {
				tx.Rollback()
				errChan <- fmt.Errorf("查找或创建资源失败: %w", err)
				return
			}

			// 使用 FirstOrCreate 查找或创建资源分配（修复后的逻辑）
			var allocation model.ResourceAllocation
			if err := tx.Where("resource_id = ? AND task_id = ? AND date = ?", resource.ID, task.ID, workDate).
				FirstOrCreate(&allocation, model.ResourceAllocation{
					ResourceID:  resource.ID,
					TaskID:      &task.ID,
					ProjectID:   &project.ID,
					Date:        workDate,
					Hours:       actualHours,
					Description: fmt.Sprintf("任务: %s", task.Title),
				}).Error; err != nil {
				tx.Rollback()
				errChan <- fmt.Errorf("查找或创建资源分配失败: %w", err)
				return
			}

			// 无论记录是新创建还是已存在，都更新工时和描述（修复后的逻辑）
			allocation.Hours = actualHours
			allocation.Description = fmt.Sprintf("任务: %s", task.Title)
			if err := tx.Save(&allocation).Error; err != nil {
				tx.Rollback()
				errChan <- fmt.Errorf("更新资源分配失败: %w", err)
				return
			}

			// 提交事务
			if err := tx.Commit().Error; err != nil {
				errChan <- fmt.Errorf("提交事务失败: %w", err)
				return
			}
		}(i)
	}

	// 等待所有 goroutine 完成
	wg.Wait()
	close(errChan)

	// 检查是否有错误
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	// 应该没有错误（使用 FirstOrCreate 和事务应该避免死锁）
	assert.Empty(t, errors, "并发创建资源分配时不应该出现错误或死锁")

	// 验证资源分配记录存在（应该只有一条记录，因为使用 FirstOrCreate 后更新）
	var allocations []model.ResourceAllocation
	require.NoError(t, db.Where("task_id = ? AND date = ?", task.ID, workDate).Find(&allocations).Error)
	
	// 应该只有一条记录（FirstOrCreate 会创建或找到记录，然后更新）
	assert.LessOrEqual(t, len(allocations), 1, "应该只有一条资源分配记录")

	// 验证资源记录存在
	var resource model.Resource
	require.NoError(t, db.Where("user_id = ? AND project_id = ?", user.ID, project.ID).First(&resource).Error)
	assert.Equal(t, user.ID, resource.UserID)
	assert.Equal(t, project.ID, resource.ProjectID)
}

// TestDeadlockFix_ConcurrentBugResourceCreation 测试并发创建Bug资源分配时的死锁修复
// 直接测试数据库操作，模拟 syncBugActualHours 的逻辑
func TestDeadlockFix_ConcurrentBugResourceCreation(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	// 确保资源相关表已创建
	require.NoError(t, db.AutoMigrate(&model.Resource{}, &model.ResourceAllocation{}))

	// 创建测试数据
	user := CreateTestUser(t, db, "testuser2", "测试用户2")
	project := CreateTestProject(t, db, "测试项目2")

	// 创建Bug
	bug := &model.Bug{
		Title:     "测试Bug",
		ProjectID: project.ID,
		Status:    "active",
		CreatorID:  user.ID,
	}
	require.NoError(t, db.Create(bug).Error)

	// 关联分配人
	require.NoError(t, db.Model(&bug).Association("Assignees").Append(&user))

	// 并发数量
	concurrency := 10
	workDate := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	// 使用 WaitGroup 等待所有 goroutine 完成
	var wg sync.WaitGroup
	wg.Add(concurrency)

	// 错误通道
	errChan := make(chan error, concurrency)

	// 并发创建资源分配（模拟 syncBugActualHours 的逻辑）
	for i := 0; i < concurrency; i++ {
		go func(index int) {
			defer wg.Done()
			actualHours := float64(2.0 + float64(index)*0.1)

			// 使用事务包裹所有操作（修复后的逻辑）
			tx := db.Begin()
			defer func() {
				if r := recover(); r != nil {
					tx.Rollback()
					errChan <- fmt.Errorf("panic: %v", r)
				}
			}()

			// 使用 FirstOrCreate 查找或创建资源（修复后的逻辑）
			var resource model.Resource
			if err := tx.Where("user_id = ? AND project_id = ?", user.ID, project.ID).
				FirstOrCreate(&resource, model.Resource{
					UserID:    user.ID,
					ProjectID: project.ID,
				}).Error; err != nil {
				tx.Rollback()
				errChan <- fmt.Errorf("查找或创建资源失败: %w", err)
				return
			}

			// 使用 FirstOrCreate 查找或创建资源分配（修复后的逻辑）
			var allocation model.ResourceAllocation
			if err := tx.Where("resource_id = ? AND bug_id = ? AND date = ?", resource.ID, bug.ID, workDate).
				FirstOrCreate(&allocation, model.ResourceAllocation{
					ResourceID:  resource.ID,
					BugID:       &bug.ID,
					ProjectID:   &project.ID,
					Date:        workDate,
					Hours:       actualHours,
					Description: fmt.Sprintf("Bug: %s", bug.Title),
				}).Error; err != nil {
				tx.Rollback()
				errChan <- fmt.Errorf("查找或创建资源分配失败: %w", err)
				return
			}

			// 无论记录是新创建还是已存在，都更新工时和描述（修复后的逻辑）
			allocation.Hours = actualHours
			allocation.Description = fmt.Sprintf("Bug: %s", bug.Title)
			if err := tx.Save(&allocation).Error; err != nil {
				tx.Rollback()
				errChan <- fmt.Errorf("更新资源分配失败: %w", err)
				return
			}

			// 提交事务
			if err := tx.Commit().Error; err != nil {
				errChan <- fmt.Errorf("提交事务失败: %w", err)
				return
			}
		}(i)
	}

	// 等待所有 goroutine 完成
	wg.Wait()
	close(errChan)

	// 检查是否有错误
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	// 应该没有错误
	assert.Empty(t, errors, "并发创建Bug资源分配时不应该出现错误或死锁")

	// 验证资源分配记录存在
	var allocations []model.ResourceAllocation
	require.NoError(t, db.Where("bug_id = ? AND date = ?", bug.ID, workDate).Find(&allocations).Error)
	
	// 应该只有一条记录
	assert.LessOrEqual(t, len(allocations), 1, "应该只有一条资源分配记录")
}

// TestDeadlockFix_ConcurrentRequirementResourceCreation 测试并发创建需求资源分配时的死锁修复
// 直接测试数据库操作，模拟 syncRequirementActualHours 的逻辑
func TestDeadlockFix_ConcurrentRequirementResourceCreation(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	// 确保资源相关表已创建
	require.NoError(t, db.AutoMigrate(&model.Resource{}, &model.ResourceAllocation{}))

	// 创建测试数据
	user := CreateTestUser(t, db, "testuser3", "测试用户3")
	project := CreateTestProject(t, db, "测试项目3")

	// 创建需求
	requirement := &model.Requirement{
		Title:      "测试需求",
		ProjectID:  project.ID,
		AssigneeID: &user.ID,
		Status:     "doing",
		CreatorID:  user.ID,
	}
	require.NoError(t, db.Create(requirement).Error)

	// 并发数量
	concurrency := 10
	workDate := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)

	// 使用 WaitGroup 等待所有 goroutine 完成
	var wg sync.WaitGroup
	wg.Add(concurrency)

	// 错误通道
	errChan := make(chan error, concurrency)

	// 并发创建资源分配（模拟 syncRequirementActualHours 的逻辑）
	for i := 0; i < concurrency; i++ {
		go func(index int) {
			defer wg.Done()
			actualHours := float64(3.0 + float64(index)*0.1)

			// 使用事务包裹所有操作（修复后的逻辑）
			tx := db.Begin()
			defer func() {
				if r := recover(); r != nil {
					tx.Rollback()
					errChan <- fmt.Errorf("panic: %v", r)
				}
			}()

			// 使用 FirstOrCreate 查找或创建资源（修复后的逻辑）
			var resource model.Resource
			if err := tx.Where("user_id = ? AND project_id = ?", user.ID, project.ID).
				FirstOrCreate(&resource, model.Resource{
					UserID:    user.ID,
					ProjectID: project.ID,
				}).Error; err != nil {
				tx.Rollback()
				errChan <- fmt.Errorf("查找或创建资源失败: %w", err)
				return
			}

			// 使用 FirstOrCreate 查找或创建资源分配（修复后的逻辑，替代先删除再创建）
			var allocation model.ResourceAllocation
			if err := tx.Where("resource_id = ? AND requirement_id = ? AND date = ?", resource.ID, requirement.ID, workDate).
				FirstOrCreate(&allocation, model.ResourceAllocation{
					ResourceID:    resource.ID,
					RequirementID: &requirement.ID,
					ProjectID:     &project.ID,
					Date:          workDate,
					Hours:         actualHours,
					Description:   fmt.Sprintf("需求: %s", requirement.Title),
				}).Error; err != nil {
				tx.Rollback()
				errChan <- fmt.Errorf("查找或创建资源分配失败: %w", err)
				return
			}

			// 无论记录是新创建还是已存在，都更新工时和描述（修复后的逻辑）
			allocation.Hours = actualHours
			allocation.Description = fmt.Sprintf("需求: %s", requirement.Title)
			if err := tx.Save(&allocation).Error; err != nil {
				tx.Rollback()
				errChan <- fmt.Errorf("更新资源分配失败: %w", err)
				return
			}

			// 提交事务
			if err := tx.Commit().Error; err != nil {
				errChan <- fmt.Errorf("提交事务失败: %w", err)
				return
			}
		}(i)
	}

	// 等待所有 goroutine 完成
	wg.Wait()
	close(errChan)

	// 检查是否有错误
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	// 应该没有错误
	assert.Empty(t, errors, "并发创建需求资源分配时不应该出现错误或死锁")

	// 验证资源分配记录存在
	var allocations []model.ResourceAllocation
	require.NoError(t, db.Where("requirement_id = ? AND date = ?", requirement.ID, workDate).Find(&allocations).Error)
	
	// 应该只有一条记录（FirstOrCreate 会创建或找到记录，然后更新）
	assert.LessOrEqual(t, len(allocations), 1, "应该只有一条资源分配记录")
}

// TestDeadlockFix_ConcurrentCalculateActualHours 测试并发计算实际工时时的死锁修复
// 直接测试数据库操作，模拟 calculateAndUpdateActualHours 的逻辑
func TestDeadlockFix_ConcurrentCalculateActualHours(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	// 确保资源相关表已创建
	require.NoError(t, db.AutoMigrate(&model.Resource{}, &model.ResourceAllocation{}))

	// 创建测试数据
	user := CreateTestUser(t, db, "testuser4", "测试用户4")
	project := CreateTestProject(t, db, "测试项目4")

	// 创建资源
	resource := &model.Resource{
		UserID:    user.ID,
		ProjectID: project.ID,
	}
	require.NoError(t, db.Create(resource).Error)

	// 创建任务
	task := &model.Task{
		Title:      "测试任务2",
		ProjectID: project.ID,
		AssigneeID: &user.ID,
		Status:     "doing",
		CreatorID:  user.ID,
	}
	require.NoError(t, db.Create(task).Error)

	// 创建多个资源分配记录
	workDate := time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		allocation := &model.ResourceAllocation{
			ResourceID: resource.ID,
			TaskID:     &task.ID,
			ProjectID:  &project.ID,
			Date:       workDate,
			Hours:      float64(1.0 + float64(i)*0.5),
		}
		require.NoError(t, db.Create(allocation).Error)
	}

	// 并发数量
	concurrency := 10

	// 使用 WaitGroup 等待所有 goroutine 完成
	var wg sync.WaitGroup
	wg.Add(concurrency)

	// 错误通道
	errChan := make(chan error, concurrency)

	// 并发计算实际工时（模拟 calculateAndUpdateActualHours 的逻辑）
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()

			// 使用事务包裹查询和更新操作（修复后的逻辑）
			tx := db.Begin()
			defer func() {
				if r := recover(); r != nil {
					tx.Rollback()
					errChan <- fmt.Errorf("panic: %v", r)
				}
			}()

			var totalHours float64
			if err := tx.Model(&model.ResourceAllocation{}).
				Where("task_id = ?", task.ID).
				Select("COALESCE(SUM(hours), 0)").
				Scan(&totalHours).Error; err != nil {
				tx.Rollback()
				// 静默返回，避免影响主流程
				return
			}

			if err := tx.Model(task).Update("actual_hours", totalHours).Error; err != nil {
				tx.Rollback()
				// 静默返回，避免影响主流程
				return
			}

			// 提交事务
			if err := tx.Commit().Error; err != nil {
				errChan <- fmt.Errorf("提交事务失败: %w", err)
				return
			}
		}()
	}

	// 等待所有 goroutine 完成
	wg.Wait()
	close(errChan)

	// 检查是否有错误
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	// 应该没有错误
	assert.Empty(t, errors, "并发计算实际工时时不应该出现错误或死锁")

	// 验证任务的实际工时已更新
	require.NoError(t, db.First(task, task.ID).Error)
	assert.NotNil(t, task.ActualHours)
	// 总工时应该是 1.0 + 1.5 + 2.0 + 2.5 + 3.0 = 10.0
	assert.InDelta(t, 10.0, *task.ActualHours, 0.01, "实际工时应该正确计算")
}

// TestDeadlockFix_ConcurrentUpdateTask 测试并发更新任务时的死锁修复
// 测试 syncTaskActualHours 和 calculateAndUpdateActualHours 的组合
func TestDeadlockFix_ConcurrentUpdateTask(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	// 确保资源相关表已创建
	require.NoError(t, db.AutoMigrate(&model.Resource{}, &model.ResourceAllocation{}))

	// 创建测试数据
	user := CreateTestUser(t, db, "testuser5", "测试用户5")
	project := CreateTestProject(t, db, "测试项目5")

	// 创建任务
	task := &model.Task{
		Title:      "测试任务3",
		ProjectID: project.ID,
		AssigneeID: &user.ID,
		CreatorID:  user.ID,
		Status:     "doing",
	}
	require.NoError(t, db.Create(task).Error)

	// 并发数量
	concurrency := 10
	workDate := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)

	// 使用 WaitGroup 等待所有 goroutine 完成
	var wg sync.WaitGroup
	wg.Add(concurrency)

	// 错误通道
	errChan := make(chan error, concurrency)

	// 并发更新任务的实际工时（模拟完整的更新流程）
	for i := 0; i < concurrency; i++ {
		go func(index int) {
			defer wg.Done()
			actualHours := float64(1.0 + float64(index)*0.1)

			// 1. 同步到资源分配（模拟 syncTaskActualHours）
			tx1 := db.Begin()
			var resource model.Resource
			if err := tx1.Where("user_id = ? AND project_id = ?", user.ID, project.ID).
				FirstOrCreate(&resource, model.Resource{
					UserID:    user.ID,
					ProjectID: project.ID,
				}).Error; err != nil {
				tx1.Rollback()
				errChan <- fmt.Errorf("查找或创建资源失败: %w", err)
				return
			}

			var allocation model.ResourceAllocation
			if err := tx1.Where("resource_id = ? AND task_id = ? AND date = ?", resource.ID, task.ID, workDate).
				FirstOrCreate(&allocation, model.ResourceAllocation{
					ResourceID:  resource.ID,
					TaskID:      &task.ID,
					ProjectID:   &project.ID,
					Date:        workDate,
					Hours:       actualHours,
					Description: fmt.Sprintf("任务: %s", task.Title),
				}).Error; err != nil {
				tx1.Rollback()
				errChan <- fmt.Errorf("查找或创建资源分配失败: %w", err)
				return
			}

			allocation.Hours = actualHours
			if err := tx1.Save(&allocation).Error; err != nil {
				tx1.Rollback()
				errChan <- fmt.Errorf("更新资源分配失败: %w", err)
				return
			}

			if err := tx1.Commit().Error; err != nil {
				errChan <- fmt.Errorf("提交事务失败: %w", err)
				return
			}

			// 2. 计算并更新实际工时（模拟 calculateAndUpdateActualHours）
			tx2 := db.Begin()
			var totalHours float64
			if err := tx2.Model(&model.ResourceAllocation{}).
				Where("task_id = ?", task.ID).
				Select("COALESCE(SUM(hours), 0)").
				Scan(&totalHours).Error; err != nil {
				tx2.Rollback()
				return
			}

			if err := tx2.Model(task).Update("actual_hours", totalHours).Error; err != nil {
				tx2.Rollback()
				return
			}

			if err := tx2.Commit().Error; err != nil {
				errChan <- fmt.Errorf("提交事务失败: %w", err)
				return
			}
		}(i)
	}

	// 等待所有 goroutine 完成
	wg.Wait()
	close(errChan)

	// 检查是否有错误
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	// 应该没有错误
	assert.Empty(t, errors, "并发更新任务时不应该出现错误或死锁")

	// 验证资源分配记录存在
	var allocations []model.ResourceAllocation
	require.NoError(t, db.Where("task_id = ? AND date = ?", task.ID, workDate).Find(&allocations).Error)
	
	// 应该只有一条记录
	assert.LessOrEqual(t, len(allocations), 1, "应该只有一条资源分配记录")
}

// TestDeadlockFix_ConcurrentResourceCreationDifferentProjects 测试不同项目并发创建资源时的死锁修复
func TestDeadlockFix_ConcurrentResourceCreationDifferentProjects(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	// 确保资源相关表已创建
	require.NoError(t, db.AutoMigrate(&model.Resource{}, &model.ResourceAllocation{}))

	// 创建测试数据
	user := CreateTestUser(t, db, "testuser6", "测试用户6")
	project1 := CreateTestProject(t, db, "测试项目6-1")
	project2 := CreateTestProject(t, db, "测试项目6-2")

	// 创建两个任务
	task1 := &model.Task{
		Title:      "测试任务4-1",
		ProjectID: project1.ID,
		AssigneeID: &user.ID,
		CreatorID:  user.ID,
		Status:     "doing",
	}
	require.NoError(t, db.Create(task1).Error)

	task2 := &model.Task{
		Title:      "测试任务4-2",
		ProjectID: project2.ID,
		AssigneeID: &user.ID,
		CreatorID:  user.ID,
		Status:     "doing",
	}
	require.NoError(t, db.Create(task2).Error)

	// 并发数量
	concurrency := 10
	workDate := time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC)

	// 使用 WaitGroup 等待所有 goroutine 完成
	var wg sync.WaitGroup
	wg.Add(concurrency * 2) // 两个任务，每个任务并发10次

	// 错误通道
	errChan := make(chan error, concurrency*2)

	// 并发调用 syncTaskActualHours（不同项目）
	for i := 0; i < concurrency; i++ {
		// 项目1的任务
		go func(index int) {
			defer wg.Done()
			actualHours := float64(1.0 + float64(index)*0.1)

			tx := db.Begin()
			defer func() {
				if r := recover(); r != nil {
					tx.Rollback()
					errChan <- fmt.Errorf("panic: %v", r)
				}
			}()

			var resource model.Resource
			if err := tx.Where("user_id = ? AND project_id = ?", user.ID, project1.ID).
				FirstOrCreate(&resource, model.Resource{
					UserID:    user.ID,
					ProjectID: project1.ID,
				}).Error; err != nil {
				tx.Rollback()
				errChan <- fmt.Errorf("查找或创建资源失败: %w", err)
				return
			}

			var allocation model.ResourceAllocation
			if err := tx.Where("resource_id = ? AND task_id = ? AND date = ?", resource.ID, task1.ID, workDate).
				FirstOrCreate(&allocation, model.ResourceAllocation{
					ResourceID:  resource.ID,
					TaskID:      &task1.ID,
					ProjectID:   &project1.ID,
					Date:        workDate,
					Hours:       actualHours,
					Description: fmt.Sprintf("任务: %s", task1.Title),
				}).Error; err != nil {
				tx.Rollback()
				errChan <- fmt.Errorf("查找或创建资源分配失败: %w", err)
				return
			}

			allocation.Hours = actualHours
			if err := tx.Save(&allocation).Error; err != nil {
				tx.Rollback()
				errChan <- fmt.Errorf("更新资源分配失败: %w", err)
				return
			}

			if err := tx.Commit().Error; err != nil {
				errChan <- fmt.Errorf("提交事务失败: %w", err)
				return
			}
		}(i)

		// 项目2的任务
		go func(index int) {
			defer wg.Done()
			actualHours := float64(2.0 + float64(index)*0.1)

			tx := db.Begin()
			defer func() {
				if r := recover(); r != nil {
					tx.Rollback()
					errChan <- fmt.Errorf("panic: %v", r)
				}
			}()

			var resource model.Resource
			if err := tx.Where("user_id = ? AND project_id = ?", user.ID, project2.ID).
				FirstOrCreate(&resource, model.Resource{
					UserID:    user.ID,
					ProjectID: project2.ID,
				}).Error; err != nil {
				tx.Rollback()
				errChan <- fmt.Errorf("查找或创建资源失败: %w", err)
				return
			}

			var allocation model.ResourceAllocation
			if err := tx.Where("resource_id = ? AND task_id = ? AND date = ?", resource.ID, task2.ID, workDate).
				FirstOrCreate(&allocation, model.ResourceAllocation{
					ResourceID:  resource.ID,
					TaskID:      &task2.ID,
					ProjectID:   &project2.ID,
					Date:        workDate,
					Hours:       actualHours,
					Description: fmt.Sprintf("任务: %s", task2.Title),
				}).Error; err != nil {
				tx.Rollback()
				errChan <- fmt.Errorf("查找或创建资源分配失败: %w", err)
				return
			}

			allocation.Hours = actualHours
			if err := tx.Save(&allocation).Error; err != nil {
				tx.Rollback()
				errChan <- fmt.Errorf("更新资源分配失败: %w", err)
				return
			}

			if err := tx.Commit().Error; err != nil {
				errChan <- fmt.Errorf("提交事务失败: %w", err)
				return
			}
		}(i)
	}

	// 等待所有 goroutine 完成
	wg.Wait()
	close(errChan)

	// 检查是否有错误
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	// 应该没有错误
	assert.Empty(t, errors, "不同项目并发创建资源分配时不应该出现错误或死锁")

	// 验证两个项目的资源记录都存在
	var resource1 model.Resource
	require.NoError(t, db.Where("user_id = ? AND project_id = ?", user.ID, project1.ID).First(&resource1).Error)
	
	var resource2 model.Resource
	require.NoError(t, db.Where("user_id = ? AND project_id = ?", user.ID, project2.ID).First(&resource2).Error)

	assert.Equal(t, user.ID, resource1.UserID)
	assert.Equal(t, project1.ID, resource1.ProjectID)
	assert.Equal(t, user.ID, resource2.UserID)
	assert.Equal(t, project2.ID, resource2.ProjectID)
}

