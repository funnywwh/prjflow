package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"project-management/internal/model"

	"gorm.io/gorm"
)

// RecordAction 记录操作（参考禅道的 create() 方法）
func RecordAction(db *gorm.DB, objectType string, objectID uint, actionType string, actorID uint, comment string, extra interface{}) (uint, error) {
	action := model.Action{
		ObjectType: objectType,
		ObjectID:   objectID,
		ActorID:    actorID,
		Action:     actionType,
		Date:       time.Now(),
		Comment:    comment,
	}

	// 获取项目ID（如果是Bug，从Bug表获取）
	if objectType == "bug" {
		var bug model.Bug
		if err := db.First(&bug, objectID).Error; err == nil {
			action.ProjectID = bug.ProjectID
		}
	}

	// 处理extra字段（JSON格式）
	if extra != nil {
		extraJSON, err := json.Marshal(extra)
		if err == nil {
			action.Extra = string(extraJSON)
		}
	}

	if err := db.Create(&action).Error; err != nil {
		return 0, err
	}

	return action.ID, nil
}

// RecordHistory 记录字段变更（参考禅道的 logHistory() 方法）
func RecordHistory(db *gorm.DB, actionID uint, changes []HistoryChange) error {
	if actionID == 0 || len(changes) == 0 {
		return nil
	}

	for _, change := range changes {
		history := model.History{
			ActionID: actionID,
			Field:    change.Field,
			Old:      change.Old,
			New:      change.New,
		}

		// 处理显示值转换
		processedHistory := ProcessHistory(db, &history)
		if err := db.Create(processedHistory).Error; err != nil {
			return err
		}
	}

	return nil
}

// HistoryChange 字段变更结构
type HistoryChange struct {
	Field string // 字段名
	Old   string // 旧值（原始值）
	New   string // 新值（原始值）
}

// ProcessHistory 处理字段值转换（参考禅道的 processHistory() 方法）
func ProcessHistory(db *gorm.DB, history *model.History) *model.History {
	// 获取操作记录以确定对象类型
	var action model.Action
	if err := db.First(&action, history.ActionID).Error; err != nil {
		return history
	}

	_ = action.ObjectType // 暂时未使用，保留用于未来扩展

	// 用户字段转换（ID转用户名）
	if isUserField(history.Field) {
		history.OldValue = getUserDisplayName(db, history.Old)
		history.NewValue = getUserDisplayName(db, history.New)
		return history
	}

	// 多选用户字段转换（如 assignee_ids）
	if isMultipleUserField(history.Field) {
		history.OldValue = getMultipleUserDisplayName(db, history.Old)
		history.NewValue = getMultipleUserDisplayName(db, history.New)
		return history
	}

	// 关联对象字段转换（ID转显示名称）
	if history.Field == "project_id" {
		history.OldValue = getProjectDisplayName(db, history.Old)
		history.NewValue = getProjectDisplayName(db, history.New)
		return history
	}
	if history.Field == "requirement_id" {
		history.OldValue = getRequirementDisplayName(db, history.Old)
		history.NewValue = getRequirementDisplayName(db, history.New)
		return history
	}
	if history.Field == "module_id" {
		history.OldValue = getModuleDisplayName(db, history.Old)
		history.NewValue = getModuleDisplayName(db, history.New)
		return history
	}
	if history.Field == "resolved_version_id" {
		history.OldValue = getVersionDisplayName(db, history.Old)
		history.NewValue = getVersionDisplayName(db, history.New)
		return history
	}

	// 枚举字段转换
	switch history.Field {
	case "status":
		history.OldValue = getStatusDisplayName(history.Old)
		history.NewValue = getStatusDisplayName(history.New)
	case "priority":
		history.OldValue = getPriorityDisplayName(history.Old)
		history.NewValue = getPriorityDisplayName(history.New)
	case "severity":
		history.OldValue = getSeverityDisplayName(history.Old)
		history.NewValue = getSeverityDisplayName(history.New)
	case "solution":
		history.OldValue = getSolutionDisplayName(history.Old)
		history.NewValue = getSolutionDisplayName(history.New)
	case "confirmed":
		history.OldValue = getBoolDisplayName(history.Old)
		history.NewValue = getBoolDisplayName(history.New)
	default:
		// 其他字段直接使用原始值
		history.OldValue = history.Old
		history.NewValue = history.New
	}

	return history
}

// CompareAndRecord 比较新旧对象并自动记录变更（参考禅道的 createChanges() 逻辑）
func CompareAndRecord(db *gorm.DB, oldObj, newObj interface{}, objectType string, objectID uint, actorID uint, actionType string) (uint, error) {
	changes := CompareObjects(oldObj, newObj)
	if len(changes) == 0 {
		return 0, nil // 没有变更，不记录
	}

	// 记录操作
	actionID, err := RecordAction(db, objectType, objectID, actionType, actorID, "", nil)
	if err != nil {
		return 0, err
	}

	// 记录字段变更
	if err := RecordHistory(db, actionID, changes); err != nil {
		return 0, err
	}

	return actionID, nil
}

// CompareObjects 比较两个对象，返回变更列表
func CompareObjects(oldObj, newObj interface{}) []HistoryChange {
	var changes []HistoryChange

	oldVal := reflect.ValueOf(oldObj)
	newVal := reflect.ValueOf(newObj)

	// 如果是指针，获取指向的值
	if oldVal.Kind() == reflect.Ptr {
		oldVal = oldVal.Elem()
	}
	if newVal.Kind() == reflect.Ptr {
		newVal = newVal.Elem()
	}

	if oldVal.Kind() != reflect.Struct || newVal.Kind() != reflect.Struct {
		return changes
	}

	oldType := oldVal.Type()
	newType := newVal.Type()

	// 遍历所有字段
	for i := 0; i < newVal.NumField(); i++ {
		newField := newType.Field(i)
		_, found := oldType.FieldByName(newField.Name)
		if !found {
			continue
		}

		// 跳过不需要记录的字段
		if shouldSkipField(newField.Name) {
			continue
		}

		newFieldVal := newVal.Field(i)
		oldFieldVal := oldVal.FieldByName(newField.Name)

		// 跳过无法比较的字段（如关联对象、切片等）
		if !isComparableField(newFieldVal) {
			continue
		}

		// 比较值
		oldStr := getFieldStringValue(oldFieldVal)
		newStr := getFieldStringValue(newFieldVal)

		if oldStr != newStr {
			changes = append(changes, HistoryChange{
				Field: getFieldJSONName(newField),
				Old:   oldStr,
				New:   newStr,
			})
		}
	}

	return changes
}

// GetFieldDisplayName 获取字段的中文显示名称
func GetFieldDisplayName(fieldName string) string {
	fieldNames := map[string]string{
		// Bug字段
		"title":              "Bug标题",
		"description":       "Bug描述",
		"status":            "Bug状态",
		"priority":          "优先级",
		"severity":          "严重程度",
		"confirmed":         "是否确认",
		"project_id":        "项目",
		"requirement_id":    "关联需求",
		"module_id":         "功能模块",
		"assignee_ids":      "指派给",
		"estimated_hours":   "预估工时",
		"actual_hours":      "实际工时",
		"solution":          "解决方案",
		"solution_note":     "解决方案备注",
		"resolved_version_id": "解决版本",
		// 需求字段
		"assignee_id":       "负责人",
		// 任务字段
		"start_date":        "开始日期",
		"end_date":          "结束日期",
		"due_date":          "截止日期",
		"progress":          "进度",
		"dependency_ids":    "依赖任务",
	}

	if name, ok := fieldNames[fieldName]; ok {
		return name
	}
	return fieldName
}

// 辅助函数

func isUserField(fieldName string) bool {
	userFields := []string{"creator_id", "assignee_id"}
	for _, f := range userFields {
		if fieldName == f {
			return true
		}
	}
	return false
}

func isMultipleUserField(fieldName string) bool {
	return fieldName == "assignee_ids"
}

func getUserDisplayName(db *gorm.DB, userIDStr string) string {
	if userIDStr == "" || userIDStr == "0" {
		return ""
	}

	var userID uint
	fmt.Sscanf(userIDStr, "%d", &userID)
	if userID == 0 {
		return ""
	}

	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		return userIDStr
	}

	if user.Nickname != "" {
		return fmt.Sprintf("%s(%s)", user.Username, user.Nickname)
	}
	return user.Username
}

func getMultipleUserDisplayName(db *gorm.DB, userIDsStr string) string {
	if userIDsStr == "" {
		return ""
	}

	// 解析ID数组（假设格式为 "[1,2,3]" 或 "1,2,3"）
	userIDsStr = strings.Trim(userIDsStr, "[]")
	if userIDsStr == "" {
		return ""
	}

	parts := strings.Split(userIDsStr, ",")
	var names []string

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		var userID uint
		fmt.Sscanf(part, "%d", &userID)
		if userID == 0 {
			continue
		}

		var user model.User
		if err := db.First(&user, userID).Error; err == nil {
			if user.Nickname != "" {
				names = append(names, fmt.Sprintf("%s(%s)", user.Username, user.Nickname))
			} else {
				names = append(names, user.Username)
			}
		}
	}

	return strings.Join(names, ",")
}

func getStatusDisplayName(status string) string {
	statusMap := map[string]string{
		"active":   "激活",
		"resolved": "已解决",
		"closed":   "已关闭",
	}
	if name, ok := statusMap[status]; ok {
		return name
	}
	return status
}

func getPriorityDisplayName(priority string) string {
	priorityMap := map[string]string{
		"low":    "低",
		"medium": "中",
		"high":   "高",
		"urgent": "紧急",
	}
	if name, ok := priorityMap[priority]; ok {
		return name
	}
	return priority
}

func getSeverityDisplayName(severity string) string {
	severityMap := map[string]string{
		"low":      "低",
		"medium":   "中",
		"high":     "高",
		"critical": "严重",
	}
	if name, ok := severityMap[severity]; ok {
		return name
	}
	return severity
}

func getSolutionDisplayName(solution string) string {
	// 解决方案已经是中文，直接返回
	return solution
}

func getBoolDisplayName(value string) string {
	if value == "true" || value == "1" {
		return "已确认"
	}
	return "未确认"
}

func shouldSkipField(fieldName string) bool {
	skipFields := []string{
		"ID", "CreatedAt", "UpdatedAt", "DeletedAt",
		"Project", "Creator", "Assignees", "Requirement", "Module", "ResolvedVersion",
		"CreatorID", // CreatorID是用户字段，有特殊处理，但不需要记录ID变更
	}
	for _, f := range skipFields {
		if fieldName == f {
			return true
		}
	}
	return false
}

func isComparableField(fieldVal reflect.Value) bool {
	kind := fieldVal.Kind()
	// 支持基本类型和指针类型
	return kind == reflect.String || kind == reflect.Int || kind == reflect.Uint ||
		kind == reflect.Float64 || kind == reflect.Bool ||
		(kind == reflect.Ptr && fieldVal.Elem().Kind() != reflect.Struct)
}

func getFieldStringValue(fieldVal reflect.Value) string {
	if !fieldVal.IsValid() {
		return ""
	}

	kind := fieldVal.Kind()

	// 处理指针类型
	if kind == reflect.Ptr {
		if fieldVal.IsNil() {
			return ""
		}
		fieldVal = fieldVal.Elem()
		kind = fieldVal.Kind()
	}

	switch kind {
	case reflect.String:
		return fieldVal.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", fieldVal.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", fieldVal.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%.2f", fieldVal.Float())
	case reflect.Bool:
		if fieldVal.Bool() {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}

func getFieldJSONName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" || jsonTag == "-" {
		return field.Name
	}

	// 处理 json tag 中的选项（如 omitempty）
	parts := strings.Split(jsonTag, ",")
	return parts[0]
}

// getProjectDisplayName 获取项目显示名称
func getProjectDisplayName(db *gorm.DB, projectIDStr string) string {
	if projectIDStr == "" || projectIDStr == "0" {
		return ""
	}

	var projectID uint
	fmt.Sscanf(projectIDStr, "%d", &projectID)
	if projectID == 0 {
		return ""
	}

	var project model.Project
	if err := db.First(&project, projectID).Error; err != nil {
		return projectIDStr
	}

	return project.Name
}

// getRequirementDisplayName 获取需求显示名称
func getRequirementDisplayName(db *gorm.DB, requirementIDStr string) string {
	if requirementIDStr == "" || requirementIDStr == "0" {
		return ""
	}

	var requirementID uint
	fmt.Sscanf(requirementIDStr, "%d", &requirementID)
	if requirementID == 0 {
		return ""
	}

	var requirement model.Requirement
	if err := db.First(&requirement, requirementID).Error; err != nil {
		return requirementIDStr
	}

	return requirement.Title
}

// getModuleDisplayName 获取模块显示名称
func getModuleDisplayName(db *gorm.DB, moduleIDStr string) string {
	if moduleIDStr == "" || moduleIDStr == "0" {
		return ""
	}

	var moduleID uint
	fmt.Sscanf(moduleIDStr, "%d", &moduleID)
	if moduleID == 0 {
		return ""
	}

	var module model.Module
	if err := db.First(&module, moduleID).Error; err != nil {
		return moduleIDStr
	}

	return module.Name
}

// getVersionDisplayName 获取版本显示名称
func getVersionDisplayName(db *gorm.DB, versionIDStr string) string {
	if versionIDStr == "" || versionIDStr == "0" {
		return ""
	}

	var versionID uint
	fmt.Sscanf(versionIDStr, "%d", &versionID)
	if versionID == 0 {
		return ""
	}

	var version model.Version
	if err := db.First(&version, versionID).Error; err != nil {
		return versionIDStr
	}

	return version.VersionNumber
}

