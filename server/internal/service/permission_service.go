package service

import (
	"errors"

	"github.com/sunmoonstrand/go-react-blog/server/internal/model"
	"gorm.io/gorm"
)

// CreatePermission 创建权限
func CreatePermission(form model.PermissionCreateForm) (int, error) {
	// 检查权限名是否已存在
	var count int64
	if err := model.DB.Model(&model.Permission{}).Where("perm_name = ?", form.PermName).Count(&count).Error; err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, errors.New("权限名已存在")
	}

	// 检查权限键是否已存在
	if form.PermKey != "" {
		if err := model.DB.Model(&model.Permission{}).Where("perm_key = ?", form.PermKey).Count(&count).Error; err != nil {
			return 0, err
		}
		if count > 0 {
			return 0, errors.New("权限键已存在")
		}
	}

	// 检查API路径是否已存在（如果是API类型）
	if form.PermType == 3 && form.APIPath != "" {
		if err := model.DB.Model(&model.Permission{}).Where("api_path = ?", form.APIPath).Count(&count).Error; err != nil {
			return 0, err
		}
		if count > 0 {
			return 0, errors.New("API路径已存在")
		}
	}

	// 创建权限
	permission := model.Permission{
		PermName:  form.PermName,
		PermKey:   form.PermKey,
		PermType:  form.PermType,
		ParentID:  form.ParentID,
		Icon:      form.Icon,
		Component: form.Component,
		Path:      form.Path,
		Redirect:  form.Redirect,
		APIPath:   form.APIPath,
		SortOrder: form.SortOrder,
		IsEnabled: form.IsEnabled,
		IsVisible: form.IsVisible,
		IsBuiltin: form.IsBuiltin,
		Remark:    form.Remark,
	}

	// 保存权限
	if err := model.DB.Create(&permission).Error; err != nil {
		return 0, err
	}

	return permission.PermID, nil
}

// UpdatePermission 更新权限
func UpdatePermission(permID int, form model.PermissionUpdateForm) error {
	// 检查权限是否存在
	var permission model.Permission
	if err := model.DB.First(&permission, permID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("权限不存在")
		}
		return err
	}

	// 检查是否为内置权限
	if permission.IsBuiltin && form.IsBuiltin != nil && !*form.IsBuiltin {
		return errors.New("内置权限不能修改为非内置权限")
	}

	// 检查权限名是否已被其他权限使用
	if form.PermName != "" && form.PermName != permission.PermName {
		var count int64
		if err := model.DB.Model(&model.Permission{}).
			Where("perm_name = ? AND perm_id != ?", form.PermName, permID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("权限名已被其他权限使用")
		}
	}

	// 检查权限键是否已被其他权限使用
	if form.PermKey != nil && *form.PermKey != "" && *form.PermKey != permission.PermKey {
		var count int64
		if err := model.DB.Model(&model.Permission{}).
			Where("perm_key = ? AND perm_id != ?", *form.PermKey, permID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("权限键已被其他权限使用")
		}
	}

	// 检查API路径是否已被其他权限使用（如果是API类型）
	if form.PermType != nil && *form.PermType == 3 && form.APIPath != nil && *form.APIPath != "" && *form.APIPath != permission.APIPath {
		var count int64
		if err := model.DB.Model(&model.Permission{}).
			Where("api_path = ? AND perm_id != ?", *form.APIPath, permID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("API路径已被其他权限使用")
		}
	}

	// 更新权限
	updates := map[string]interface{}{}
	if form.PermName != "" {
		updates["perm_name"] = form.PermName
	}
	if form.PermKey != nil {
		updates["perm_key"] = *form.PermKey
	}
	if form.PermType != nil {
		updates["perm_type"] = *form.PermType
	}
	if form.ParentID != nil {
		updates["parent_id"] = *form.ParentID
	}
	if form.Icon != nil {
		updates["icon"] = *form.Icon
	}
	if form.Component != nil {
		updates["component"] = *form.Component
	}
	if form.Path != nil {
		updates["path"] = *form.Path
	}
	if form.Redirect != nil {
		updates["redirect"] = *form.Redirect
	}
	if form.APIPath != nil {
		updates["api_path"] = *form.APIPath
	}
	if form.SortOrder != nil {
		updates["sort_order"] = *form.SortOrder
	}
	if form.IsEnabled != nil {
		updates["is_enabled"] = *form.IsEnabled
	}
	if form.IsVisible != nil {
		updates["is_visible"] = *form.IsVisible
	}
	if form.IsBuiltin != nil {
		updates["is_builtin"] = *form.IsBuiltin
	}
	if form.Remark != nil {
		updates["remark"] = *form.Remark
	}

	if err := model.DB.Model(&permission).Updates(updates).Error; err != nil {
		return err
	}

	return nil
}

// GetPermissionByID 根据ID获取权限
func GetPermissionByID(permID int) (*model.Permission, error) {
	var permission model.Permission
	if err := model.DB.First(&permission, permID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("权限不存在")
		}
		return nil, err
	}
	return &permission, nil
}

// ListPermissions 获取权限列表
func ListPermissions(params model.PermissionQueryParams) (*model.PageResult, error) {
	var permissions []model.Permission
	var total int64

	// 构建查询
	query := model.DB.Model(&model.Permission{})

	// 应用过滤条件
	if params.PermName != "" {
		query = query.Where("perm_name LIKE ?", "%"+params.PermName+"%")
	}
	if params.PermKey != "" {
		query = query.Where("perm_key LIKE ?", "%"+params.PermKey+"%")
	}
	if params.PermType != nil {
		query = query.Where("perm_type = ?", *params.PermType)
	}
	if params.ParentID != nil {
		query = query.Where("parent_id = ?", *params.ParentID)
	}
	if params.IsEnabled != nil {
		query = query.Where("is_enabled = ?", *params.IsEnabled)
	}
	if params.IsVisible != nil {
		query = query.Where("is_visible = ?", *params.IsVisible)
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
	if err := query.Offset(offset).Limit(params.PageSize).Order("sort_order ASC, perm_id ASC").Find(&permissions).Error; err != nil {
		return nil, err
	}

	// 转换为响应对象
	var permissionResponses []model.PermissionResponse
	for _, perm := range permissions {
		permissionResponses = append(permissionResponses, model.PermissionResponse{
			PermID:    perm.PermID,
			PermName:  perm.PermName,
			PermKey:   perm.PermKey,
			PermType:  perm.PermType,
			ParentID:  perm.ParentID,
			Icon:      perm.Icon,
			Component: perm.Component,
			Path:      perm.Path,
			Redirect:  perm.Redirect,
			APIPath:   perm.APIPath,
			SortOrder: perm.SortOrder,
			IsEnabled: perm.IsEnabled,
			IsVisible: perm.IsVisible,
			IsBuiltin: perm.IsBuiltin,
			Remark:    perm.Remark,
			CreatedAt: perm.CreatedAt,
			UpdatedAt: perm.UpdatedAt,
		})
	}

	return model.NewPageResult(permissionResponses, total, params.Page, params.PageSize), nil
}

// DeletePermission 删除权限
func DeletePermission(permID int) error {
	// 检查权限是否存在
	var permission model.Permission
	if err := model.DB.First(&permission, permID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("权限不存在")
		}
		return err
	}

	// 检查是否为内置权限
	if permission.IsBuiltin {
		return errors.New("内置权限不能删除")
	}

	// 检查是否有子权限
	var count int64
	if err := model.DB.Model(&model.Permission{}).Where("parent_id = ?", permID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("该权限下有子权限，不能删除")
	}

	// 检查权限是否已分配给角色
	if err := model.DB.Model(&model.RolePermission{}).Where("perm_id = ?", permID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("权限已分配给角色，不能删除")
	}

	// 删除权限
	if err := model.DB.Delete(&permission).Error; err != nil {
		return err
	}

	return nil
}

// GetAllPermissions 获取所有权限（用于下拉选择）
func GetAllPermissions() ([]model.Option, error) {
	var permissions []model.Permission
	if err := model.DB.Where("is_enabled = ?", true).Order("sort_order ASC, perm_id ASC").Find(&permissions).Error; err != nil {
		return nil, err
	}

	var options []model.Option
	for _, perm := range permissions {
		options = append(options, model.Option{
			Label: perm.PermName,
			Value: perm.PermID,
		})
	}

	return options, nil
}

// GetPermissionTree 获取权限树
func GetPermissionTree() ([]*model.Permission, error) {
	var permissions []model.Permission
	if err := model.DB.Order("sort_order ASC, perm_id ASC").Find(&permissions).Error; err != nil {
		return nil, err
	}

	return BuildPermissionTree(permissions), nil
}

// UpdatePermissionStatus 更新权限状态
func UpdatePermissionStatus(permID int, isEnabled bool) error {
	// 检查权限是否存在
	var permission model.Permission
	if err := model.DB.First(&permission, permID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("权限不存在")
		}
		return err
	}

	// 更新状态
	if err := model.DB.Model(&permission).Update("is_enabled", isEnabled).Error; err != nil {
		return err
	}

	return nil
}
