package service

import (
	"errors"
	"strconv"

	"github.com/sunmoonstrand/go-react-blog/server/internal/model"
	"gorm.io/gorm"
)

// CreateConfig 创建配置
func CreateConfig(form model.ConfigCreateForm) (int, error) {
	// 检查配置键是否已存在
	var count int64
	if err := model.DB.Model(&model.SysConfig{}).Where("config_key = ?", form.ConfigKey).Count(&count).Error; err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, errors.New("配置键已存在")
	}

	// 创建配置
	config := model.SysConfig{
		ConfigName:  form.ConfigName,
		ConfigKey:   form.ConfigKey,
		ConfigValue: form.ConfigValue,
		ValueType:   form.ValueType,
		ConfigGroup: form.ConfigGroup,
		IsBuiltin:   form.IsBuiltin,
		IsFrontend:  form.IsFrontend,
		SortOrder:   form.SortOrder,
		Remark:      form.Remark,
	}

	// 保存配置
	if err := model.DB.Create(&config).Error; err != nil {
		return 0, err
	}

	return config.ConfigID, nil
}

// UpdateConfig 更新配置
func UpdateConfig(configID int, form model.ConfigUpdateForm) error {
	// 检查配置是否存在
	var config model.SysConfig
	if err := model.DB.First(&config, configID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("配置不存在")
		}
		return err
	}

	// 检查是否为内置配置
	if config.IsBuiltin {
		return errors.New("内置配置不能修改")
	}

	// 更新配置
	updates := map[string]interface{}{}
	if form.ConfigName != "" {
		updates["config_name"] = form.ConfigName
	}
	if form.ConfigValue != "" {
		updates["config_value"] = form.ConfigValue
	}
	if form.ValueType != 0 {
		updates["value_type"] = form.ValueType
	}
	if form.ConfigGroup != "" {
		updates["config_group"] = form.ConfigGroup
	}
	updates["is_frontend"] = form.IsFrontend
	if form.SortOrder != 0 {
		updates["sort_order"] = form.SortOrder
	}
	if form.Remark != "" {
		updates["remark"] = form.Remark
	}

	if err := model.DB.Model(&config).Updates(updates).Error; err != nil {
		return err
	}

	return nil
}

// GetConfigByID 根据ID获取配置
func GetConfigByID(configID int) (*model.SysConfig, error) {
	var config model.SysConfig
	if err := model.DB.First(&config, configID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("配置不存在")
		}
		return nil, err
	}
	return &config, nil
}

// GetConfigByKey 根据键获取配置
func GetConfigByKey(configKey string) (*model.SysConfig, error) {
	var config model.SysConfig
	if err := model.DB.Where("config_key = ?", configKey).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("配置不存在")
		}
		return nil, err
	}
	return &config, nil
}

// ListConfigs 获取配置列表
func ListConfigs(params model.ConfigQueryParams) (*model.PageResult, error) {
	var configs []model.SysConfig
	var total int64

	// 构建查询
	query := model.DB.Model(&model.SysConfig{})

	// 应用过滤条件
	if params.ConfigGroup != "" {
		query = query.Where("config_group = ?", params.ConfigGroup)
	}
	if params.ConfigKey != "" {
		query = query.Where("config_key LIKE ?", "%"+params.ConfigKey+"%")
	}
	if params.Keyword != "" {
		query = query.Where("config_name LIKE ? OR config_key LIKE ? OR config_value LIKE ?",
			"%"+params.Keyword+"%", "%"+params.Keyword+"%", "%"+params.Keyword+"%")
	}
	if params.IsFrontend != nil {
		query = query.Where("is_frontend = ?", *params.IsFrontend)
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
	if err := query.Offset(offset).Limit(params.PageSize).Order("config_group, sort_order ASC, config_id ASC").Find(&configs).Error; err != nil {
		return nil, err
	}

	// 转换为响应对象
	var configResponses []model.ConfigResponse
	for _, config := range configs {
		configResponses = append(configResponses, model.ConfigResponse{
			ConfigID:    config.ConfigID,
			ConfigName:  config.ConfigName,
			ConfigKey:   config.ConfigKey,
			ConfigValue: config.ConfigValue,
			ValueType:   config.ValueType,
			ConfigGroup: config.ConfigGroup,
			IsBuiltin:   config.IsBuiltin,
			IsFrontend:  config.IsFrontend,
			SortOrder:   config.SortOrder,
			Remark:      config.Remark,
			CreatedAt:   config.CreatedAt,
			UpdatedAt:   config.UpdatedAt,
		})
	}

	return model.NewPageResult(configResponses, total, params.Page, params.PageSize), nil
}

// DeleteConfig 删除配置
func DeleteConfig(configID int) error {
	// 检查配置是否存在
	var config model.SysConfig
	if err := model.DB.First(&config, configID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("配置不存在")
		}
		return err
	}

	// 检查是否为内置配置
	if config.IsBuiltin {
		return errors.New("内置配置不能删除")
	}

	// 删除配置
	if err := model.DB.Delete(&config).Error; err != nil {
		return err
	}

	return nil
}

// GetFrontendConfigs 获取前端配置
func GetFrontendConfigs() (map[string]interface{}, error) {
	var configs []model.SysConfig
	if err := model.DB.Where("is_frontend = ?", true).Find(&configs).Error; err != nil {
		return nil, err
	}

	// 转换为键值对
	result := make(map[string]interface{})
	for _, config := range configs {
		// 根据值类型转换
		switch config.ValueType {
		case 1: // 字符串
			result[config.ConfigKey] = config.ConfigValue
		case 2: // 整数
			if val, err := strconv.Atoi(config.ConfigValue); err == nil {
				result[config.ConfigKey] = val
			} else {
				result[config.ConfigKey] = 0
			}
		case 3: // 浮点数
			if val, err := strconv.ParseFloat(config.ConfigValue, 64); err == nil {
				result[config.ConfigKey] = val
			} else {
				result[config.ConfigKey] = 0.0
			}
		case 4: // 布尔值
			if val, err := strconv.ParseBool(config.ConfigValue); err == nil {
				result[config.ConfigKey] = val
			} else {
				result[config.ConfigKey] = false
			}
		default:
			result[config.ConfigKey] = config.ConfigValue
		}
	}

	return result, nil
}

// GetConfigGroups 获取配置分组
func GetConfigGroups() ([]string, error) {
	var groups []string
	if err := model.DB.Model(&model.SysConfig{}).
		Select("DISTINCT config_group").
		Pluck("config_group", &groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}
