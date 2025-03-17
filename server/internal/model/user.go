package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	UserID         int       `gorm:"column:user_id;primaryKey;autoIncrement" json:"user_id"`
	Username       string    `gorm:"column:username;size:30;not null;unique" json:"username"`
	PasswordHash   string    `gorm:"column:password_hash;size:100" json:"-"`
	Email          string    `gorm:"column:email;size:100;unique" json:"email"`
	Mobile         string    `gorm:"column:mobile;size:20;unique" json:"mobile"`
	WechatOpenID   string    `gorm:"column:wechat_openid;size:50;unique" json:"-"`
	WechatUnionID  string    `gorm:"column:wechat_unionid;size:50;unique" json:"-"`
	Avatar         string    `gorm:"column:avatar;size:255" json:"avatar"`
	Nickname       string    `gorm:"column:nickname;size:50" json:"nickname"`
	RealName       string    `gorm:"column:real_name;size:50" json:"real_name"`
	Gender         int8      `gorm:"column:gender;default:0" json:"gender"`
	Birthday       time.Time `gorm:"column:birthday" json:"birthday"`
	Status         int8      `gorm:"column:status;not null;default:1" json:"status"`
	RegisterSource int8      `gorm:"column:register_source;not null;default:1" json:"register_source"`
	LastLogin      time.Time `gorm:"column:last_login" json:"last_login"`
	LoginCount     int       `gorm:"column:login_count;not null;default:0" json:"login_count"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	Roles          []Role    `gorm:"many2many:sys_user_roles;foreignKey:UserID;joinForeignKey:UserID;References:RoleID;joinReferences:RoleID" json:"roles"`
}

// TableName 指定表名
func (User) TableName() string {
	return "sys_users"
}

// BeforeCreate 创建前的钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// 如果密码不为空，则加密密码
	if u.PasswordHash != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.PasswordHash), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.PasswordHash = string(hashedPassword)
	}
	return nil
}

// BeforeUpdate 更新前的钩子
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	// 如果密码不为空，则加密密码
	if u.PasswordHash != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.PasswordHash), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.PasswordHash = string(hashedPassword)
	}
	return nil
}

// CheckPassword 检查密码是否正确
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// UserLoginForm 用户登录表单
type UserLoginForm struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"123456"`
}

// UserRegisterForm 用户注册表单
type UserRegisterForm struct {
	Username string `json:"username" binding:"required,min=4,max=30" example:"newuser"`
	Password string `json:"password" binding:"required,min=6,max=20" example:"123456"`
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Mobile   string `json:"mobile" binding:"omitempty,len=11" example:"13800138000"`
	Nickname string `json:"nickname" binding:"omitempty,max=50" example:"新用户"`
}

// UserUpdateForm 用户信息更新表单
type UserUpdateForm struct {
	Nickname string    `json:"nickname" binding:"omitempty,max=50" example:"新昵称"`
	Avatar   string    `json:"avatar" binding:"omitempty,max=255" example:"http://example.com/avatar.jpg"`
	Gender   int8      `json:"gender" binding:"omitempty,oneof=0 1 2" example:"1"`
	Birthday time.Time `json:"birthday" binding:"omitempty" example:"1990-01-01T00:00:00Z"`
	Email    string    `json:"email" binding:"omitempty,email" example:"newemail@example.com"`
	Mobile   string    `json:"mobile" binding:"omitempty,len=11" example:"13900139000"`
}

// PasswordUpdateForm 密码更新表单
type PasswordUpdateForm struct {
	OldPassword string `json:"old_password" binding:"required" example:"123456"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=20" example:"654321"`
}

// UserResponse 用户信息响应
type UserResponse struct {
	UserID         int       `json:"user_id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	Mobile         string    `json:"mobile"`
	Avatar         string    `json:"avatar"`
	Nickname       string    `json:"nickname"`
	Gender         int8      `json:"gender"`
	Status         int8      `json:"status"`
	RegisterSource int8      `json:"register_source"`
	LastLogin      time.Time `json:"last_login"`
	CreatedAt      time.Time `json:"created_at"`
	Roles          []string  `json:"roles"`
}
