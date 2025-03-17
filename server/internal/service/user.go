package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sunmoonstrand/go-react-blog/server/config"
	"github.com/sunmoonstrand/go-react-blog/server/internal/model"
	"github.com/sunmoonstrand/go-react-blog/server/internal/utils/jwt_utils"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RegisterUser 注册新用户
func RegisterUser(form model.UserCreateForm) (int, error) {
	// 检查用户名是否已存在
	var count int64
	if err := model.DB.Model(&model.User{}).Where("username = ?", form.Username).Count(&count).Error; err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	if form.Email != "" {
		if err := model.DB.Model(&model.User{}).Where("email = ?", form.Email).Count(&count).Error; err != nil {
			return 0, err
		}
		if count > 0 {
			return 0, errors.New("邮箱已被使用")
		}
	}

	// 检查手机号是否已存在
	if form.Mobile != "" {
		if err := model.DB.Model(&model.User{}).Where("mobile = ?", form.Mobile).Count(&count).Error; err != nil {
			return 0, err
		}
		if count > 0 {
			return 0, errors.New("手机号已被使用")
		}
	}

	// 创建用户
	user := model.User{
		Username:   form.Username,
		Nickname:   form.Nickname,
		Password:   form.Password, // 密码会在BeforeCreate钩子中加密
		Email:      form.Email,
		Mobile:     form.Mobile,
		Avatar:     form.Avatar,
		Gender:     form.Gender,
		IsEnabled:  form.IsEnabled,
		IsVerified: form.IsVerified,
		Remark:     form.Remark,
	}

	// 保存用户
	if err := model.DB.Create(&user).Error; err != nil {
		return 0, err
	}

	// 分配默认角色（普通用户角色，假设ID为2）
	if err := model.DB.Exec("INSERT INTO sys_user_roles (user_id, role_id) VALUES (?, ?)", user.UserID, 2).Error; err != nil {
		zap.L().Error("分配默认角色失败",
			zap.Int("user_id", user.UserID),
			zap.Error(err),
		)
		// 不返回错误，因为用户已创建成功
	}

	return user.UserID, nil
}

// LoginUser 用户登录
func LoginUser(username, password string) (*model.LoginResult, error) {
	var user model.User

	// 查询用户
	if err := model.DB.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 检查用户状态
	if !user.IsEnabled {
		return nil, errors.New("账号已被禁用")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("密码错误")
	}

	// 查询用户角色
	var roleIDs []int
	if err := model.DB.Table("sys_user_roles").
		Select("role_id").
		Where("user_id = ?", user.UserID).
		Pluck("role_id", &roleIDs).Error; err != nil {
		return nil, err
	}

	// 生成访问令牌
	accessToken, err := jwt_utils.GenerateToken(user.UserID, roleIDs, config.AppConfig.JWT.AccessTokenExpire)
	if err != nil {
		return nil, err
	}

	// 生成刷新令牌
	refreshToken, err := jwt_utils.GenerateRefreshToken(user.UserID, config.AppConfig.JWT.RefreshTokenExpire)
	if err != nil {
		return nil, err
	}

	// 更新最后登录时间
	if err := model.DB.Model(&user).Update("last_login_at", time.Now()).Error; err != nil {
		zap.L().Error("更新最后登录时间失败",
			zap.Int("user_id", user.UserID),
			zap.Error(err),
		)
		// 不返回错误，因为登录已成功
	}

	return &model.LoginResult{
		UserID:       user.UserID,
		Username:     user.Username,
		Nickname:     user.Nickname,
		Avatar:       user.Avatar,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    config.AppConfig.JWT.AccessTokenExpire,
	}, nil
}

// RefreshToken 刷新访问令牌
func RefreshToken(refreshToken string) (*model.RefreshTokenResult, error) {
	// 解析刷新令牌
	claims, err := jwt_utils.ParseRefreshToken(refreshToken)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("刷新令牌已过期")
		}
		return nil, errors.New("无效的刷新令牌")
	}

	// 验证用户是否存在
	var user model.User
	if err := model.DB.First(&user, claims.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 检查用户状态
	if !user.IsEnabled {
		return nil, errors.New("账号已被禁用")
	}

	// 查询用户角色
	var roleIDs []int
	if err := model.DB.Table("sys_user_roles").
		Select("role_id").
		Where("user_id = ?", user.UserID).
		Pluck("role_id", &roleIDs).Error; err != nil {
		return nil, err
	}

	// 生成新的访问令牌
	accessToken, err := jwt_utils.GenerateToken(user.UserID, roleIDs, config.AppConfig.JWT.AccessTokenExpire)
	if err != nil {
		return nil, err
	}

	return &model.RefreshTokenResult{
		AccessToken: accessToken,
		ExpiresIn:   config.AppConfig.JWT.AccessTokenExpire,
	}, nil
}

// GetUserByID 根据ID获取用户信息
func GetUserByID(userID int) (*model.User, error) {
	var user model.User
	if err := model.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserRoles 获取用户角色
func GetUserRoles(userID int) ([]model.Role, error) {
	var roles []model.Role
	if err := model.DB.Table("sys_roles").
		Select("sys_roles.*").
		Joins("JOIN sys_user_roles ON sys_roles.role_id = sys_user_roles.role_id").
		Where("sys_user_roles.user_id = ?", userID).
		Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// UpdateUserProfile 更新用户个人资料
func UpdateUserProfile(userID int, form model.UserProfileUpdateForm) error {
	// 检查邮箱是否已被其他用户使用
	if form.Email != "" {
		var count int64
		if err := model.DB.Model(&model.User{}).
			Where("email = ? AND user_id != ?", form.Email, userID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("邮箱已被其他用户使用")
		}
	}

	// 检查手机号是否已被其他用户使用
	if form.Mobile != "" {
		var count int64
		if err := model.DB.Model(&model.User{}).
			Where("mobile = ? AND user_id != ?", form.Mobile, userID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("手机号已被其他用户使用")
		}
	}

	// 更新用户资料
	updates := map[string]interface{}{
		"nickname": form.Nickname,
		"email":    form.Email,
		"mobile":   form.Mobile,
		"avatar":   form.Avatar,
		"gender":   form.Gender,
	}

	if err := model.DB.Model(&model.User{}).Where("user_id = ?", userID).Updates(updates).Error; err != nil {
		return err
	}

	return nil
}

// ChangePassword 修改密码
func ChangePassword(userID int, oldPassword, newPassword string) error {
	var user model.User
	if err := model.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("原密码错误")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 更新密码
	if err := model.DB.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		return err
	}

	return nil
}

// ResetPassword 重置密码（管理员操作）
func ResetPassword(userID int, newPassword string) error {
	var user model.User
	if err := model.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 更新密码
	if err := model.DB.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		return err
	}

	return nil
}

// ListUsers 获取用户列表
func ListUsers(params model.UserQueryParams) (*model.PageResult, error) {
	var users []model.User
	var total int64

	// 构建查询
	query := model.DB.Model(&model.User{})

	// 应用过滤条件
	if params.Username != "" {
		query = query.Where("username LIKE ?", "%"+params.Username+"%")
	}
	if params.Nickname != "" {
		query = query.Where("nickname LIKE ?", "%"+params.Nickname+"%")
	}
	if params.Email != "" {
		query = query.Where("email LIKE ?", "%"+params.Email+"%")
	}
	if params.Mobile != "" {
		query = query.Where("mobile LIKE ?", "%"+params.Mobile+"%")
	}
	if params.IsEnabled != nil {
		query = query.Where("is_enabled = ?", *params.IsEnabled)
	}
	if params.IsVerified != nil {
		query = query.Where("is_verified = ?", *params.IsVerified)
	}
	if params.Gender != nil {
		query = query.Where("gender = ?", *params.Gender)
	}
	if params.StartTime != nil && params.EndTime != nil {
		query = query.Where("created_at BETWEEN ? AND ?", params.StartTime, params.EndTime)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	if err := query.Offset(offset).Limit(params.PageSize).Order("user_id DESC").Find(&users).Error; err != nil {
		return nil, err
	}

	// 转换为响应对象
	var userResponses []model.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, model.UserResponse{
			UserID:      user.UserID,
			Username:    user.Username,
			Nickname:    user.Nickname,
			Email:       user.Email,
			Mobile:      user.Mobile,
			Avatar:      user.Avatar,
			Gender:      user.Gender,
			IsEnabled:   user.IsEnabled,
			IsVerified:  user.IsVerified,
			LastLoginAt: user.LastLoginAt,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		})
	}

	return model.NewPageResult(userResponses, total, params.Page, params.PageSize), nil
}

// UpdateUserStatus 更新用户状态
func UpdateUserStatus(userID int, isEnabled bool) error {
	if err := model.DB.Model(&model.User{}).Where("user_id = ?", userID).Update("is_enabled", isEnabled).Error; err != nil {
		return err
	}
	return nil
}

// DeleteUser 删除用户
func DeleteUser(userID int) error {
	// 开启事务
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除用户角色关联
	if err := tx.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 删除用户
	if err := tx.Delete(&model.User{}, userID).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// AssignRoles 分配用户角色
func AssignRoles(userID int, roleIDs []int) error {
	// 开启事务
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除现有角色
	if err := tx.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 分配新角色
	for _, roleID := range roleIDs {
		userRole := model.UserRole{
			UserID: userID,
			RoleID: roleID,
		}
		if err := tx.Create(&userRole).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}
