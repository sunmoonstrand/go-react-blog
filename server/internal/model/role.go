package model

import (
	"time"
)

// Role 角色模型
type Role struct {
	RoleID      int          `gorm:"column:role_id;primaryKey;autoIncrement" json:"role_id"`
	RoleName    string       `gorm:"column:role_name;size:50;not null;unique" json:"role_name"`
	RoleKey     string       `gorm:"column:role_key;size:50;not null;unique" json:"role_key"`
	RoleSort    int16        `gorm:"column:role_sort;not null;default:0" json:"role_sort"`
	RoleDesc    string       `gorm:"column:role_desc;size:200" json:"role_desc"`
	IsDefault   bool         `gorm:"column:is_default;not null;default:false" json:"is_default"`
	IsEnabled   bool         `gorm:"column:is_enabled;not null;default:true" json:"is_enabled"`
	CreatedAt   time.Time    `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	Permissions []Permission `gorm:"many2many:sys_role_permissions;foreignKey:RoleID;joinForeignKey:RoleID;References:PermID;joinReferences:PermID" json:"permissions"`
	Users       []User       `gorm:"many2many:sys_user_roles;foreignKey:RoleID;joinForeignKey:RoleID;References:UserID;joinReferences:UserID" json:"users"`
}

// TableName 指定表名
func (Role) TableName() string {
	return "sys_roles"
}

// RoleCreateForm 角色创建表单
type RoleCreateForm struct {
	RoleName  string `json:"role_name" binding:"required,max=50" example:"编辑角色"`
	RoleKey   string `json:"role_key" binding:"required,max=50" example:"editor"`
	RoleSort  int16  `json:"role_sort" binding:"omitempty" example:"5"`
	RoleDesc  string `json:"role_desc" binding:"omitempty,max=200" example:"负责内容编辑的角色"`
	IsDefault bool   `json:"is_default" example:"false"`
	IsEnabled bool   `json:"is_enabled" example:"true"`
}

// RoleUpdateForm 角色更新表单
type RoleUpdateForm struct {
	RoleName  string `json:"role_name" binding:"omitempty,max=50" example:"编辑角色"`
	RoleKey   string `json:"role_key" binding:"omitempty,max=50" example:"editor"`
	RoleSort  int16  `json:"role_sort" binding:"omitempty" example:"5"`
	RoleDesc  string `json:"role_desc" binding:"omitempty,max=200" example:"负责内容编辑的角色"`
	IsDefault bool   `json:"is_default" example:"false"`
	IsEnabled bool   `json:"is_enabled" example:"true"`
}

// RolePermissionForm 角色权限分配表单
type RolePermissionForm struct {
	PermissionIDs []int `json:"permission_ids" binding:"required" example:"1,2,3,4"`
}

// RoleResponse 角色信息响应
type RoleResponse struct {
	RoleID      int       `json:"role_id"`
	RoleName    string    `json:"role_name"`
	RoleKey     string    `json:"role_key"`
	RoleSort    int16     `json:"role_sort"`
	RoleDesc    string    `json:"role_desc"`
	IsDefault   bool      `json:"is_default"`
	IsEnabled   bool      `json:"is_enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Permissions []string  `json:"permissions,omitempty"`
}
