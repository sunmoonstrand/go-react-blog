package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sunmoonstrand/go-react-blog/server/internal/model"
)

// GenerateToken 生成JWT令牌
func GenerateToken(userID int, username string, roleIDs []int, secret string, expire int, issuer string) (string, error) {
	// 创建JWT声明
	claims := model.CustomClaims{
		UserID:   userID,
		Username: username,
		RoleIDs:  roleIDs,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	return token.SignedString([]byte(secret))
}

// GenerateRefreshToken 生成刷新令牌
func GenerateRefreshToken(userID int, secret string, expire int, issuer string) (string, error) {
	// 创建JWT声明
	claims := model.RefreshClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	return token.SignedString([]byte(secret))
}

// ParseToken 解析JWT令牌
func ParseToken(tokenString string, secret string) (*model.CustomClaims, error) {
	// 创建声明
	claims := &model.CustomClaims{}

	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名算法")
		}
		return []byte(secret), nil
	})

	// 处理解析错误
	if err != nil {
		return nil, err
	}

	// 验证令牌有效性
	if !token.Valid {
		return nil, errors.New("无效的令牌")
	}

	return claims, nil
}

// ParseRefreshToken 解析刷新令牌
func ParseRefreshToken(tokenString string, secret string) (*model.RefreshClaims, error) {
	// 创建声明
	claims := &model.RefreshClaims{}

	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名算法")
		}
		return []byte(secret), nil
	})

	// 处理解析错误
	if err != nil {
		return nil, err
	}

	// 验证令牌有效性
	if !token.Valid {
		return nil, errors.New("无效的令牌")
	}

	return claims, nil
}
