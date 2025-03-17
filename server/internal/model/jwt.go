package model

import (
	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims 自定义JWT声明
type CustomClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	RoleIDs  []int  `json:"role_ids"`
	jwt.RegisteredClaims
}

// RefreshClaims 刷新令牌声明
type RefreshClaims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}
