package service

import (
	"strings"

	"github.com/sunmoonstrand/go-react-blog/server/internal/model"
	"go.uber.org/zap"
)

// CheckPermission 检查用户是否有权限访问指定路径
func CheckPermission(userID int, roleIDs []int, requestPath, requestMethod string) (bool, error) {
	// 超级管理员拥有所有权限
	for _, roleID := range roleIDs {
		if roleID == 1 {
			return true, nil
		}
	}

	// 查询用户角色对应的权限
	var permissions []model.Permission
	result := model.DB.Table("sys_permissions").
		Select("sys_permissions.*").
		Joins("JOIN sys_role_permissions ON sys_permissions.perm_id = sys_role_permissions.perm_id").
		Where("sys_role_permissions.role_id IN ?", roleIDs).
		Where("sys_permissions.is_enabled = ?", true).
		Find(&permissions)

	if result.Error != nil {
		zap.L().Error("查询用户权限失败",
			zap.Int("user_id", userID),
			zap.Any("role_ids", roleIDs),
			zap.Error(result.Error),
		)
		return false, result.Error
	}

	// 检查是否有匹配的API权限
	for _, perm := range permissions {
		// 只检查API类型的权限（perm_type=3）
		if perm.PermType == 3 && perm.APIPath != "" {
			// 检查API路径是否匹配
			if matchAPIPath(perm.APIPath, requestPath, requestMethod) {
				return true, nil
			}
		}
	}

	return false, nil
}

// matchAPIPath 检查API路径是否匹配
func matchAPIPath(permPath, requestPath, requestMethod string) bool {
	// 分割权限路径，格式为：METHOD:/path/to/resource
	parts := strings.SplitN(permPath, ":", 2)
	if len(parts) != 2 {
		return false
	}

	// 检查请求方法是否匹配
	permMethod := strings.ToUpper(parts[0])
	if permMethod != "*" && permMethod != strings.ToUpper(requestMethod) {
		return false
	}

	// 检查路径是否匹配
	permPathPattern := parts[1]

	// 处理通配符
	if strings.HasSuffix(permPathPattern, "/*") {
		// 去掉末尾的/*
		prefix := permPathPattern[:len(permPathPattern)-2]
		return strings.HasPrefix(requestPath, prefix)
	}

	// 精确匹配
	return permPathPattern == requestPath
}

// GetUserPermissions 获取用户权限列表
func GetUserPermissions(userID int) ([]model.Permission, error) {
	var permissions []model.Permission

	// 查询用户角色
	var userRoles []model.Role
	if err := model.DB.Preload("Permissions").
		Joins("JOIN sys_user_roles ON sys_roles.role_id = sys_user_roles.role_id").
		Where("sys_user_roles.user_id = ?", userID).
		Where("sys_roles.is_enabled = ?", true).
		Find(&userRoles).Error; err != nil {
		return nil, err
	}

	// 收集所有权限
	permMap := make(map[int]model.Permission)
	for _, role := range userRoles {
		for _, perm := range role.Permissions {
			if perm.IsEnabled {
				permMap[perm.PermID] = perm
			}
		}
	}

	// 转换为切片
	for _, perm := range permMap {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// BuildPermissionTree 构建权限树
func BuildPermissionTree(permissions []model.Permission) []*model.Permission {
	// 创建权限映射
	permMap := make(map[int]*model.Permission)
	for i := range permissions {
		perm := permissions[i]
		permMap[perm.PermID] = &perm
	}

	// 构建树结构
	var rootNodes []*model.Permission
	for _, perm := range permMap {
		if perm.ParentID == nil || *perm.ParentID == 0 {
			// 根节点
			rootNodes = append(rootNodes, perm)
		} else {
			// 子节点
			if parent, ok := permMap[*perm.ParentID]; ok {
				if parent.Children == nil {
					parent.Children = []*model.Permission{}
				}
				parent.Children = append(parent.Children, perm)
			}
		}
	}

	return rootNodes
}
