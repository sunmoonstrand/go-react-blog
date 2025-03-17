package model

import (
	"time"
)

// Permission 权限模型
type Permission struct {
	PermID    int           `gorm:"column:perm_id;primaryKey;autoIncrement" json:"perm_id"`
	PermName  string        `gorm:"column:perm_name;size:50;not null" json:"perm_name"`
	PermKey   string        `gorm:"column:perm_key;size:50;not null;unique" json:"perm_key"`
	PermType  int8          `gorm:"column:perm_type;not null;default:1" json:"perm_type"`
	ParentID  *int          `gorm:"column:parent_id" json:"parent_id"`
	Path      string        `gorm:"column:path" json:"path"`
	APIPath   string        `gorm:"column:api_path;size:200" json:"api_path"`
	Component string        `gorm:"column:component;size:100" json:"component"`
	Perms     string        `gorm:"column:perms;size:100" json:"perms"`
	Icon      string        `gorm:"column:icon;size:100" json:"icon"`
	MenuSort  int16         `gorm:"column:menu_sort;not null;default:0" json:"menu_sort"`
	IsVisible bool          `gorm:"column:is_visible;not null;default:true" json:"is_visible"`
	IsEnabled bool          `gorm:"column:is_enabled;not null;default:true" json:"is_enabled"`
	CreatedAt time.Time     `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time     `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	Roles     []Role        `gorm:"many2many:sys_role_permissions;foreignKey:PermID;joinForeignKey:PermID;References:RoleID;joinReferences:RoleID" json:"roles"`
	Children  []*Permission `gorm:"-" json:"children,omitempty"`
}

// TableName 指定表名
func (Permission) TableName() string {
	return "sys_permissions"
}

// PermissionCreateForm 权限创建表单
type PermissionCreateForm struct {
	PermName  string `json:"perm_name" binding:"required,max=50" example:"用户管理"`
	PermKey   string `json:"perm_key" binding:"required,max=50" example:"system:user:list"`
	PermType  int8   `json:"perm_type" binding:"required,oneof=1 2 3" example:"1"`
	ParentID  *int   `json:"parent_id" example:"0"`
	APIPath   string `json:"api_path" binding:"omitempty,max=200" example:"/api/v1/users"`
	Component string `json:"component" binding:"omitempty,max=100" example:"system/user/index"`
	Perms     string `json:"perms" binding:"omitempty,max=100" example:"system:user:list"`
	Icon      string `json:"icon" binding:"omitempty,max=100" example:"user"`
	MenuSort  int16  `json:"menu_sort" binding:"omitempty" example:"1"`
	IsVisible bool   `json:"is_visible" example:"true"`
	IsEnabled bool   `json:"is_enabled" example:"true"`
}

// PermissionUpdateForm 权限更新表单
type PermissionUpdateForm struct {
	PermName  string `json:"perm_name" binding:"omitempty,max=50" example:"用户管理"`
	PermKey   string `json:"perm_key" binding:"omitempty,max=50" example:"system:user:list"`
	PermType  int8   `json:"perm_type" binding:"omitempty,oneof=1 2 3" example:"1"`
	ParentID  *int   `json:"parent_id" example:"0"`
	APIPath   string `json:"api_path" binding:"omitempty,max=200" example:"/api/v1/users"`
	Component string `json:"component" binding:"omitempty,max=100" example:"system/user/index"`
	Perms     string `json:"perms" binding:"omitempty,max=100" example:"system:user:list"`
	Icon      string `json:"icon" binding:"omitempty,max=100" example:"user"`
	MenuSort  int16  `json:"menu_sort" binding:"omitempty" example:"1"`
	IsVisible bool   `json:"is_visible" example:"true"`
	IsEnabled bool   `json:"is_enabled" example:"true"`
}

// PermissionResponse 权限信息响应
type PermissionResponse struct {
	PermID    int                   `json:"perm_id"`
	PermName  string                `json:"perm_name"`
	PermKey   string                `json:"perm_key"`
	PermType  int8                  `json:"perm_type"`
	ParentID  *int                  `json:"parent_id"`
	Path      string                `json:"path"`
	APIPath   string                `json:"api_path"`
	Component string                `json:"component"`
	Perms     string                `json:"perms"`
	Icon      string                `json:"icon"`
	MenuSort  int16                 `json:"menu_sort"`
	IsVisible bool                  `json:"is_visible"`
	IsEnabled bool                  `json:"is_enabled"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	Children  []*PermissionResponse `json:"children,omitempty"`
}
