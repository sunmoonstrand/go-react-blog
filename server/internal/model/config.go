package model

import (
	"time"
)

// SysConfig 系统配置模型
type SysConfig struct {
	ConfigID    int       `gorm:"column:config_id;primaryKey;autoIncrement" json:"config_id"`
	ConfigName  string    `gorm:"column:config_name;size:100;not null" json:"config_name"`
	ConfigKey   string    `gorm:"column:config_key;size:100;not null;unique" json:"config_key"`
	ConfigValue string    `gorm:"column:config_value;not null" json:"config_value"`
	ValueType   int8      `gorm:"column:value_type;not null;default:1" json:"value_type"`
	ConfigGroup string    `gorm:"column:config_group;size:50;not null;default:'default'" json:"config_group"`
	IsBuiltin   bool      `gorm:"column:is_builtin;not null;default:false" json:"is_builtin"`
	IsFrontend  bool      `gorm:"column:is_frontend;not null;default:false" json:"is_frontend"`
	SortOrder   int16     `gorm:"column:sort_order;not null;default:0" json:"sort_order"`
	Remark      string    `gorm:"column:remark;size:200" json:"remark"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName 指定表名
func (SysConfig) TableName() string {
	return "sys_configs"
}

// ConfigCreateForm 配置创建表单
type ConfigCreateForm struct {
	ConfigName  string `json:"config_name" binding:"required,max=100" example:"网站名称"`
	ConfigKey   string `json:"config_key" binding:"required,max=100" example:"site_name"`
	ConfigValue string `json:"config_value" binding:"required" example:"我的博客"`
	ValueType   int8   `json:"value_type" binding:"required,oneof=1 2 3 4" example:"1"`
	ConfigGroup string `json:"config_group" binding:"required,max=50" example:"site"`
	IsBuiltin   bool   `json:"is_builtin" example:"false"`
	IsFrontend  bool   `json:"is_frontend" example:"true"`
	SortOrder   int16  `json:"sort_order" binding:"omitempty" example:"1"`
	Remark      string `json:"remark" binding:"omitempty,max=200" example:"网站名称配置"`
}

// ConfigUpdateForm 配置更新表单
type ConfigUpdateForm struct {
	ConfigName  string `json:"config_name" binding:"omitempty,max=100" example:"网站名称"`
	ConfigValue string `json:"config_value" binding:"required" example:"我的博客"`
	ValueType   int8   `json:"value_type" binding:"omitempty,oneof=1 2 3 4" example:"1"`
	ConfigGroup string `json:"config_group" binding:"omitempty,max=50" example:"site"`
	IsFrontend  bool   `json:"is_frontend" example:"true"`
	SortOrder   int16  `json:"sort_order" binding:"omitempty" example:"1"`
	Remark      string `json:"remark" binding:"omitempty,max=200" example:"网站名称配置"`
}

// ConfigQueryParams 配置查询参数
type ConfigQueryParams struct {
	ConfigGroup string `form:"config_group" json:"config_group"`
	ConfigKey   string `form:"config_key" json:"config_key"`
	Keyword     string `form:"keyword" json:"keyword"`
	IsFrontend  *bool  `form:"is_frontend" json:"is_frontend"`
	IsBuiltin   *bool  `form:"is_builtin" json:"is_builtin"`
	Page        int    `form:"page" json:"page" binding:"required,min=1" default:"1"`
	PageSize    int    `form:"page_size" json:"page_size" binding:"required,min=1,max=100" default:"10"`
}

// ConfigResponse 配置信息响应
type ConfigResponse struct {
	ConfigID    int       `json:"config_id"`
	ConfigName  string    `json:"config_name"`
	ConfigKey   string    `json:"config_key"`
	ConfigValue string    `json:"config_value"`
	ValueType   int8      `json:"value_type"`
	ConfigGroup string    `json:"config_group"`
	IsBuiltin   bool      `json:"is_builtin"`
	IsFrontend  bool      `json:"is_frontend"`
	SortOrder   int16     `json:"sort_order"`
	Remark      string    `json:"remark"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
