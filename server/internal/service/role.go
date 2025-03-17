package service

import (
	"errors"

	"github.com/sunmoonstrand/go-react-blog/server/internal/model"
	"gorm.io/gorm"
)

// CreateRole 创建角色
func CreateRole(form model.RoleCreateForm) (int, error) {
	// 检查角色名是否已存在
	var count int64
	if err := model.DB.Model(&model.Role{}).Where("role_name = ?", form.RoleName).Count(&count).Error; err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, errors.New("角色名已存在")
	}

	// 检查角色键是否已存在
	if err := model.DB.Model(&model.Role{}).Where("role_key = ?", form.RoleKey).Count(&count).Error; err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, errors.New("角色键已存在")
	}

	// 创建角色
	role := model.Role{
		RoleName:  form.RoleName,
		RoleKey:   form.RoleKey,
		SortOrder: form.SortOrder,
		IsEnabled: form.IsEnabled,
		IsBuiltin: form.IsBuiltin,
		Remark:    form.Remark,
	}

	// 保存角色
	if err := model.DB.Create(&role).Error; err != nil {
		return 0, err
	}

	return role.RoleID, nil
}

// UpdateRole 更新角色
func UpdateRole(roleID int, form model.RoleUpdateForm) error {
	// 检查角色是否存在
	var role model.Role
	if err := model.DB.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("角色不存在")
		}
		return err
	}

	// 检查是否为内置角色
	if role.IsBuiltin && form.IsBuiltin != nil && !*form.IsBuiltin {
		return errors.New("内置角色不能修改为非内置角色")
	}

	// 检查角色名是否已被其他角色使用
	if form.RoleName != "" && form.RoleName != role.RoleName {
		var count int64
		if err := model.DB.Model(&model.Role{}).
			Where("role_name = ? AND role_id != ?", form.RoleName, roleID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("角色名已被其他角色使用")
		}
	}

	// 检查角色键是否已被其他角色使用
	if form.RoleKey != "" && form.RoleKey != role.RoleKey {
		var count int64
		if err := model.DB.Model(&model.Role{}).
			Where("role_key = ? AND role_id != ?", form.RoleKey, roleID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("角色键已被其他角色使用")
		}
	}

	// 更新角色
	updates := map[string]interface{}{}
	if form.RoleName != "" {
		updates["role_name"] = form.RoleName
	}
	if form.RoleKey != "" {
		updates["role_key"] = form.RoleKey
	}
	if form.SortOrder != nil {
		updates["sort_order"] = *form.SortOrder
	}
	if form.IsEnabled != nil {
		updates["is_enabled"] = *form.IsEnabled
	}
	if form.IsBuiltin != nil {
		updates["is_builtin"] = *form.IsBuiltin
	}
	if form.Remark != nil {
		updates["remark"] = *form.Remark
	}

	if err := model.DB.Model(&role).Updates(updates).Error; err != nil {
		return err
	}

	return nil
}

// GetRoleByID 根据ID获取角色
func GetRoleByID(roleID int) (*model.Role, error) {
	var role model.Role
	if err := model.DB.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("角色不存在")
		}
		return nil, err
	}
	return &role, nil
}

// ListRoles 获取角色列表
func ListRoles(params model.RoleQueryParams) (*model.PageResult, error) {
	var roles []model.Role
	var total int64

	// 构建查询
	query := model.DB.Model(&model.Role{})

	// 应用过滤条件
	if params.RoleName != "" {
		query = query.Where("role_name LIKE ?", "%"+params.RoleName+"%")
	}
	if params.RoleKey != "" {
		query = query.Where("role_key LIKE ?", "%"+params.RoleKey+"%")
	}
	if params.IsEnabled != nil {
		query = query.Where("is_enabled = ?", *params.IsEnabled)
	}
	if params.IsBuiltin != nil {
		query = query.Where("is_builtin = ?", *params.IsBuiltin)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	if err := query.Offset(offset).Limit(params.PageSize).Order("sort_order ASC, role_id ASC").Find(&roles).Error; err != nil {
		return nil, err
	}

	// 转换为响应对象
	var roleResponses []model.RoleResponse
	for _, role := range roles {
		roleResponses = append(roleResponses, model.RoleResponse{
			RoleID:    role.RoleID,
			RoleName:  role.RoleName,
			RoleKey:   role.RoleKey,
			SortOrder: role.SortOrder,
			IsEnabled: role.IsEnabled,
			IsBuiltin: role.IsBuiltin,
			Remark:    role.Remark,
			CreatedAt: role.CreatedAt,
			UpdatedAt: role.UpdatedAt,
		})
	}

	return model.NewPageResult(roleResponses, total, params.Page, params.PageSize), nil
}

// DeleteRole 删除角色
func DeleteRole(roleID int) error {
	// 检查角色是否存在
	var role model.Role
	if err := model.DB.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("角色不存在")
		}
		return err
	}

	// 检查是否为内置角色
	if role.IsBuiltin {
		return errors.New("内置角色不能删除")
	}

	// 检查角色是否已分配给用户
	var count int64
	if err := model.DB.Model(&model.UserRole{}).Where("role_id = ?", roleID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("角色已分配给用户，不能删除")
	}

	// 开启事务
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除角色权限关联
	if err := tx.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 删除角色
	if err := tx.Delete(&role).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetAllRoles 获取所有角色（用于下拉选择）
func GetAllRoles() ([]model.Option, error) {
	var roles []model.Role
	if err := model.DB.Where("is_enabled = ?", true).Order("sort_order ASC, role_id ASC").Find(&roles).Error; err != nil {
		return nil, err
	}

	var options []model.Option
	for _, role := range roles {
		options = append(options, model.Option{
			Label: role.RoleName,
			Value: role.RoleID,
		})
	}

	return options, nil
}

// AssignPermissions 分配角色权限
func AssignPermissions(roleID int, permIDs []int) error {
	// 检查角色是否存在
	var role model.Role
	if err := model.DB.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("角色不存在")
		}
		return err
	}

	// 开启事务
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除现有权限
	if err := tx.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 分配新权限
	for _, permID := range permIDs {
		rolePermission := model.RolePermission{
			RoleID: roleID,
			PermID: permID,
		}
		if err := tx.Create(&rolePermission).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// GetRolePermissions 获取角色权限
func GetRolePermissions(roleID int) ([]int, error) {
	var permIDs []int
	if err := model.DB.Table("sys_role_permissions").
		Select("perm_id").
		Where("role_id = ?", roleID).
		Pluck("perm_id", &permIDs).Error; err != nil {
		return nil, err
	}
	return permIDs, nil
}

// UpdateRoleStatus 更新角色状态
func UpdateRoleStatus(roleID int, isEnabled bool) error {
	// 检查角色是否存在
	var role model.Role
	if err := model.DB.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("角色不存在")
		}
		return err
	}

	// 更新状态
	if err := model.DB.Model(&role).Update("is_enabled", isEnabled).Error; err != nil {
		return err
	}

	return nil
}
