package v1

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/sunmoonstrand/go-react-blog/server/internal/auth"
	"github.com/sunmoonstrand/go-react-blog/server/internal/logger"
	"github.com/sunmoonstrand/go-react-blog/server/internal/model"
	"github.com/sunmoonstrand/go-react-blog/server/internal/utils/resp"
)

// AuthController 认证控制器
type AuthController struct {
	userModel     *model.UserModel
	loginLogModel *model.LoginLogModel
	jwtService    *auth.JWTService
}

// NewAuthController 创建认证控制器实例
func NewAuthController(userModel *model.UserModel, loginLogModel *model.LoginLogModel, jwtService *auth.JWTService) *AuthController {
	return &AuthController{
		userModel:     userModel,
		loginLogModel: loginLogModel,
		jwtService:    jwtService,
	}
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=4,max=30"`
	Password string `json:"password" binding:"required,min=6,max=20"`
}

// RegisterRequest 注册请求结构
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=4,max=30"`
	Password string `json:"password" binding:"required,min=6,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Nickname string `json:"nickname" binding:"required,min=2,max=30"`
	Mobile   string `json:"mobile" binding:"omitempty,len=11"`
}

// RefreshTokenRequest 刷新令牌请求结构
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Login 用户登录
// @Summary 用户登录
// @Description 使用用户名和密码登录系统
// @Tags 认证管理
// @Accept json
// @Produce json
// @Param data body LoginRequest true "登录信息"
// @Success 200 {object} resp.Response 成功返回用户信息和token
// @Failure 400 {object} resp.Response 请求参数错误
// @Failure 401 {object} resp.Response 用户名或密码错误
// @Failure 500 {object} resp.Response 服务器内部错误
// @Router /api/v1/auth/login [post]
func (ac *AuthController) Login(c *gin.Context) {
	var loginReq LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		resp.FailWithValidation(c, err)
		return
	}

	// 获取用户信息
	user, err := ac.userModel.GetUserByUsername(loginReq.Username)
	if err != nil {
		logger.Error("用户登录失败", "username", loginReq.Username, "error", err)
		resp.FailWithMsg(c, "用户名或密码错误")
		return
	}

	// 检查用户状态
	if user.Status != 1 {
		// 记录登录日志
		go ac.recordLoginLog(user.ID, loginReq.Username, 2, 2, c, "账号状态异常")
		resp.FailWithMsg(c, "账号已被禁用或未激活")
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(loginReq.Password)); err != nil {
		// 记录登录日志
		go ac.recordLoginLog(user.ID, loginReq.Username, 1, 2, c, "密码错误")
		resp.FailWithMsg(c, "用户名或密码错误")
		return
	}

	// 生成JWT令牌
	token, refreshToken, err := ac.jwtService.GenerateTokens(user.ID, user.Username)
	if err != nil {
		logger.Error("生成JWT令牌失败", "user_id", user.ID, "error", err)
		resp.FailWithMsg(c, "登录失败，请稍后重试")
		return
	}

	// 更新用户最后登录时间和登录次数
	go ac.userModel.UpdateLoginInfo(user.ID)

	// 记录登录日志
	go ac.recordLoginLog(user.ID, user.Username, 1, 1, c, "")

	// 返回登录成功信息
	resp.OkWithData(c, gin.H{
		"access_token":  token,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"user_info": gin.H{
			"user_id":  user.ID,
			"username": user.Username,
			"nickname": user.Nickname,
			"avatar":   user.Avatar,
			"email":    user.Email,
		},
	})
}

// Register 用户注册
// @Summary 用户注册
// @Description 注册新用户
// @Tags 认证管理
// @Accept json
// @Produce json
// @Param data body RegisterRequest true "注册信息"
// @Success 200 {object} resp.Response 注册成功
// @Failure 400 {object} resp.Response 请求参数错误
// @Failure 409 {object} resp.Response 用户名已存在
// @Failure 500 {object} resp.Response 服务器内部错误
// @Router /api/v1/auth/register [post]
func (ac *AuthController) Register(c *gin.Context) {
	var registerReq RegisterRequest
	if err := c.ShouldBindJSON(&registerReq); err != nil {
		resp.FailWithValidation(c, err)
		return
	}

	// 检查用户名是否已存在
	exists, err := ac.userModel.CheckUsernameExists(registerReq.Username)
	if err != nil {
		logger.Error("检查用户名是否存在失败", "username", registerReq.Username, "error", err)
		resp.FailWithMsg(c, "注册失败，请稍后重试")
		return
	}
	if exists {
		resp.FailWithCode(c, http.StatusConflict, "用户名已被使用")
		return
	}

	// 检查邮箱是否已存在
	exists, err = ac.userModel.CheckEmailExists(registerReq.Email)
	if err != nil {
		logger.Error("检查邮箱是否存在失败", "email", registerReq.Email, "error", err)
		resp.FailWithMsg(c, "注册失败，请稍后重试")
		return
	}
	if exists {
		resp.FailWithCode(c, http.StatusConflict, "邮箱已被注册")
		return
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerReq.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("密码加密失败", "error", err)
		resp.FailWithMsg(c, "注册失败，请稍后重试")
		return
	}

	// 创建用户
	user := &model.User{
		Username:       registerReq.Username,
		PasswordHash:   string(hashedPassword),
		Email:          registerReq.Email,
		Nickname:       registerReq.Nickname,
		Mobile:         registerReq.Mobile,
		Status:         1, // 正常状态
		RegisterSource: 1, // 邮箱注册
	}

	if err := ac.userModel.CreateUser(user); err != nil {
		logger.Error("创建用户失败", "error", err)
		resp.FailWithMsg(c, "注册失败，请稍后重试")
		return
	}

	// 分配默认角色
	if err := ac.userModel.AssignDefaultRoles(user.ID); err != nil {
		logger.Error("分配默认角色失败", "user_id", user.ID, "error", err)
		// 这里不返回错误，因为用户已创建成功，角色分配失败可以后续修复
	}

	resp.OkWithMsg(c, "注册成功")
}

// RefreshToken 刷新令牌
// @Summary 刷新令牌
// @Description 使用refresh_token刷新访问令牌
// @Tags 认证管理
// @Accept json
// @Produce json
// @Param data body RefreshTokenRequest true "刷新令牌"
// @Success 200 {object} resp.Response 刷新成功，返回新的令牌
// @Failure 400 {object} resp.Response 请求参数错误
// @Failure 401 {object} resp.Response 刷新令牌无效或已过期
// @Failure 500 {object} resp.Response 服务器内部错误
// @Router /api/v1/auth/refresh [post]
func (ac *AuthController) RefreshToken(c *gin.Context) {
	var refreshReq RefreshTokenRequest
	if err := c.ShouldBindJSON(&refreshReq); err != nil {
		resp.FailWithValidation(c, err)
		return
	}

	// 验证刷新令牌
	claims, err := ac.jwtService.ValidateRefreshToken(refreshReq.RefreshToken)
	if err != nil {
		resp.FailWithCode(c, http.StatusUnauthorized, "刷新令牌无效或已过期")
		return
	}

	// 生成新的令牌
	token, refreshToken, err := ac.jwtService.GenerateTokens(claims.UserID, claims.Username)
	if err != nil {
		logger.Error("生成新JWT令牌失败", "user_id", claims.UserID, "error", err)
		resp.FailWithMsg(c, "刷新令牌失败，请重新登录")
		return
	}

	resp.OkWithData(c, gin.H{
		"access_token":  token,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
	})
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户退出登录
// @Tags 认证管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} resp.Response 登出成功
// @Failure 401 {object} resp.Response 未授权
// @Router /api/v1/auth/logout [post]
func (ac *AuthController) Logout(c *gin.Context) {
	// 从JWT中获取用户信息
	userID, ok := c.Get("user_id")
	if !ok {
		resp.FailWithCode(c, http.StatusUnauthorized, "未授权")
		return
	}

	// 将当前令牌加入黑名单
	token := c.GetHeader("Authorization")
	if token != "" && len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
		// 将令牌加入黑名单
		if err := ac.jwtService.BlacklistToken(token); err != nil {
			logger.Error("将令牌加入黑名单失败", "user_id", userID, "error", err)
			// 这里不返回错误，因为即使黑名单失败，前端也会清除令牌
		}
	}

	resp.OkWithMsg(c, "登出成功")
}

// recordLoginLog 记录登录日志
func (ac *AuthController) recordLoginLog(userID uint, username string, loginType, loginStatus uint8, c *gin.Context, failReason string) {
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	loginLog := &model.LoginLog{
		UserID:      userID,
		Username:    username,
		LoginType:   loginType,
		LoginStatus: loginStatus,
		IPAddress:   clientIP,
		UserAgent:   userAgent,
		LoginTime:   time.Now(),
		FailReason:  failReason,
	}

	if err := ac.loginLogModel.CreateLoginLog(loginLog); err != nil {
		logger.Error("记录登录日志失败", "user_id", userID, "error", err)
	}
}

// RegisterRoutes 注册路由
func (ac *AuthController) RegisterRoutes(router *gin.RouterGroup) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", ac.Login)
		authGroup.POST("/register", ac.Register)
		authGroup.POST("/refresh", ac.RefreshToken)
		authGroup.POST("/logout", ac.Logout)
	}
}
