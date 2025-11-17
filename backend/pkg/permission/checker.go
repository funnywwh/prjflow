package permission

import (
	"gorm.io/gorm"
	"project-management/internal/model"
)

// CheckPermission 检查用户角色是否有权限
func CheckPermission(roles []string, permCode string) bool {
	// 这里应该从数据库查询角色权限
	// 为了简化，这里先返回true，实际实现需要查询数据库
	return true
}

// GetRolePermissions 获取角色的所有权限
func GetRolePermissions(db *gorm.DB, roleCodes []string) ([]string, error) {
	var permissions []model.Permission

	err := db.Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN roles ON role_permissions.role_id = roles.id").
		Where("roles.code IN ?", roleCodes).
		Find(&permissions).Error

	if err != nil {
		return nil, err
	}

	permCodes := make([]string, 0, len(permissions))
	for _, perm := range permissions {
		permCodes = append(permCodes, perm.Code)
	}

	return permCodes, nil
}

// CheckPermissionWithDB 使用数据库检查权限
func CheckPermissionWithDB(db *gorm.DB, roleCodes []string, permCode string) (bool, error) {
	permissions, err := GetRolePermissions(db, roleCodes)
	if err != nil {
		return false, err
	}

	for _, perm := range permissions {
		if perm == permCode {
			return true, nil
		}
	}

	return false, nil
}

