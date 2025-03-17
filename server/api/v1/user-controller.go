package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/sunmoonstrand/go-react-blog/server/internal/logger"
	"github.com/sunmoonstrand/go-react-blog/server/internal/middleware"
	"github.com/sunmoonstrand/go-react-blog/server/internal/model"
	"github.com/sunmoonstrand/go-react-blog/server/internal/utils/resp"
)

// UserController 用户控制器
type UserController struct {
	userModel *model.UserModel
	roleModel *model.RoleModel
}

// NewUserController 创建用户控制器实例
func NewUserController(userModel *model.UserModel, roleModel *model.RoleModel) *UserController {
	return &UserController{
		userModel: userModel,
		roleModel: roleModel,
	}
}

// CreateUserRequest 创建用户请求结构
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=4,max=30"`
	Password string `json:"password" binding:"required,min=6,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Nickname string `json:"nickname" binding:"required,min=2,max=30"`
	Mobile   string `json:"mobile" binding:"omitempty,len=11"`
	Status   uint8  `json:"status" binding:"required,oneof=1 2 3"`
	RoleIDs  []uint `json:"role_ids" binding:"required,min=1"`
}

// UpdateUserRequest 更新用户请求结构
type UpdateUserRequest struct {
	Email    string `json:"email" binding:"omitempty,email"`
	Nickname string `json:"nickname" binding:"omitempty,min=2,max=30"`
	Mobile   string `json:"mobile" binding:"omitempty,len=11"`
	Avatar   string `json:"avatar"`
	Status   uint8  `json:"status" binding:"omitempty,oneof=1 2 3"`
	RoleIDs  []uint `json:"role_ids" binding:"omitempty,min=1"`
}

// UpdatePasswordRequest 更新密码请求结构
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=6,max=20"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=20,nefield=OldPassword"`
}

// GetUserList 获取用户列表
// @Summary 获取用户列表
// @Description 分页获取用户列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码，默认1" default(1)
// @Param page_size query int false "每页记录数，默认10" default(10)
// @Param username query string false "用户名筛选"
// @Param status query int false "状态筛选:1正常,2禁用,3未激活"
// @Success 200 {object} resp.Response{data=resp.PageResult} "返回用户列表"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/users [get]
func (uc *UserController) GetUserList(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 获取筛选参数
	username := c.Query("username")
	status, _ := strconv.Atoi(c.DefaultQuery("status", "0"))

	// 调用服务获取用户列表
	users, total, err := uc.userModel.GetUserList(page, pageSize, username, uint8(status))
	if err != nil {
		logger.Error("获取用户列表失败", "error", err)
		resp.FailWithMsg(c, "获取用户列表失败")
		return
	}

	// 返回结果
	resp.OkWithPage(c, users, total, page, pageSize)
}

// GetUserByID 根据ID获取用户信息
// @Summary 获取用户详情
// @Description 根据用户ID获取用户详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "用户ID"
// @Success 200 {object} resp.Response "返回用户详情"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 404 {object} resp.Response "用户不存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/users/{id} [get]
func (uc *UserController) GetUserByID(c *gin.Context) {
	// 获取路径参数
	userID, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil {
		resp.FailWithMsg(c, "无效的用户ID")
		return
	}

	// 获取用户信息
	user, err := uc.userModel.GetUserByID(uint(userID))
	if err != nil {
		if err.Error() == "record not found" {
			resp.FailWithCode(c, http.StatusNotFound, "用户不存在")
		} else {
			logger.Error("获取用户信息失败", "user_id", userID, "error", err)
			resp.FailWithMsg(c, "获取用户信息失败")
		}
		return
	}

	// 获取用户角色
	roles, err := uc.userModel.GetUserRoles(uint(userID))
	if err != nil {
		logger.Error("获取用户角色失败", "user_id", userID, "error", err)
		// 这里不返回错误，只是角色可能为空
	}

	// 构建响应数据
	respData := gin.H{
		"user_id":     user.ID,
		"username":    user.Username,
		"email":       user.Email,
		"mobile":      user.Mobile,
		"nickname":    user.Nickname,
		"avatar":      user.Avatar,
		"status":      user.Status,
		"gender":      user.Gender,
		"created_at":  user.CreatedAt,
		"last_login":  user.LastLogin,
		"login_count": user.LoginCount,
		"roles":       roles,
	}

	resp.OkWithData(c, respData)
}

// CreateUser 创建用户
// @Summary 创建用户
// @Description 管理员创建新用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body CreateUserRequest true "用户信息"
// @Success 200 {object} resp.Response "创建成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 409 {object} resp.Response "用户名或邮箱已存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/users [post]
func (uc *UserController) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.FailWithValidation(c, err)
		return
	}

	// 检查用户名是否已存在
	exists, err := uc.userModel.CheckUsernameExists(req.Username)
	if err != nil {
		logger.Error("检查用户名是否存在失败", "username", req.Username, "error", err)
		resp.FailWithMsg(c, "创建用户失败，请稍后重试")
		return
	}
	if exists {
		resp.FailWithCode(c, http.StatusConflict, "用户名已被使用")
		return
	}

	// 检查邮箱是否已存在
	exists, err = uc.userModel.CheckEmailExists(req.Email)
	if err != nil {
		logger.Error("检查邮箱是否存在失败", "email", req.Email, "error", err)
		resp.FailWithMsg(c, "创建用户失败，请稍后重试")
		return
	}
	if exists {
		resp.FailWithCode(c, http.StatusConflict, "邮箱已被注册")
		return
	}

	// 检查角色是否存在
	for _, roleID := range req.RoleIDs {
		exists, err := uc.roleModel.CheckRoleExists(roleID)
		if err != nil {
			logger.Error("检查角色是否存在失败", "role_id", roleID, "error", err)
			resp.FailWithMsg(c, "创建用户失败，请稍后重试")
			return
		}
		if !exists {
			resp.FailWithMsg(c, "指定的角色不存在")
			return
		}
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("密码加密失败", "error", err)
		resp.FailWithMsg(c, "创建用户失败，请稍后重试")
		return
	}

	// 创建用户
	user := &model.User{
		Username:       req.Username,
		PasswordHash:   string(hashedPassword),
		Email:          req.Email,
		Nickname:       req.Nickname,
		Mobile:         req.Mobile,
		Status:         req.Status,
		RegisterSource: 1, // 后台创建
	}

	if err := uc.userModel.CreateUser(user); err != nil {
		logger.Error("创建用户失败", "error", err)
		resp.FailWithMsg(c, "创建用户失败，请稍后重试")
		return
	}

	// 分配角色
	if err := uc.userModel.AssignRoles(user.ID, req.RoleIDs); err != nil {
		logger.Error("分配角色失败", "user_id", user.ID, "error", err)
		resp.FailWithMsg(c, "用户创建成功，但角色分配失败")
		return
	}

	resp.OkWithMsg(c, "创建用户成功")
}

// UpdateUser 更新用户信息
// @Summary 更新用户信息
// @Description 更新用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "用户ID"
// @Param data body UpdateUserRequest true "用户信息"
// @Success 200 {object} resp.Response "更新成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 404 {object} resp.Response "用户不存在"
// @Failure 409 {object} resp.Response "邮箱已被其他用户使用"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/users/{id} [put]
func (uc *UserController) UpdateUser(c *gin.Context) {
	// 获取路径参数
	userID, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil {
		resp.FailWithMsg(c, "无效的用户ID")
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.FailWithValidation(c, err)
		return
	}

	// 检查用户是否存在
	exists, err := uc.userModel.CheckUserExists(uint(userID))
	if err != nil {
		logger.Error("检查用户是否存在失败", "user_id", userID, "error", err)
		resp.FailWithMsg(c, "更新用户失败，请稍后重试")
		return
	}
	if !exists {
		resp.FailWithCode(c, http.StatusNotFound, "用户不存在")
		return
	}

	// 如果更新邮箱，检查邮箱是否被其他用户使用
	if req.Email != "" {
		emailExists, err := uc.userModel.CheckEmailExistsExcept(req.Email, uint(userID))
		if err != nil {
			logger.Error("检查邮箱是否存在失败", "email", req.Email, "error", err)
			resp.FailWithMsg(c, "更新用户失败，请稍后重试")
			return
		}
		if emailExists {
			resp.FailWithCode(c, http.StatusConflict, "邮箱已被其他用户使用")
			return
		}
	}

	// 如果更新角色，检查角色是否存在
	if len(req.RoleIDs) > 0 {
		for _, roleID := range req.RoleIDs {
			exists, err := uc.roleModel.CheckRoleExists(roleID)
			if err != nil {
				logger.Error("检查角色是否存在失败", "role_id", roleID, "error", err)
				resp.FailWithMsg(c, "更新用户失败，请稍后重试")
				return
			}
			if !exists {
				resp.FailWithMsg(c, "指定的角色不存在")
				return
			}
		}
	}

	// 构建用户更新信息
	updates := make(map[string]interface{})
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Mobile != "" {
		updates["mobile"] = req.Mobile
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Status != 0 {
		updates["status"] = req.Status
	}

	// 更新用户信息
	if len(updates) > 0 {
		if err := uc.userModel.UpdateUser(uint(userID), updates); err != nil {
			logger.Error("更新用户信息失败", "user_id", userID, "error", err)
			resp.FailWithMsg(c, "更新用户信息失败")
			return
		}
	}

	// 更新用户角色
	if len(req.RoleIDs) > 0 {
		if err := uc.userModel.UpdateUserRoles(uint(userID), req.RoleIDs); err != nil {
			logger.Error("更新用户角色失败", "user_id", userID, "error", err)
			resp.FailWithMsg(c, "用户信息更新成功，但角色更新失败")
			return
		}
	}

	resp.OkWithMsg(c, "更新用户信息成功")
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 根据ID删除用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "用户ID"
// @Success 200 {object} resp.Response "删除成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 404 {object} resp.Response "用户不存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/users/{id} [delete]
func (uc *UserController) DeleteUser(c *gin.Context) {
	// 获取路径参数
	userID, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil {
		resp.FailWithMsg(c, "无效的用户ID")
		return
	}

	// 获取当前用户ID
	currentUserID, exists := c.Get("user_id")
	if !exists {
		resp.FailWithCode(c, http.StatusUnauthorized, "未授权")
		return
	}

	// 不能删除自己
	if uint(userID) == currentUserID.(uint) {
		resp.FailWithMsg(c, "不能删除当前登录用户")
		return
	}

	// 检查用户是否存在
	exists, err = uc.userModel.CheckUserExists(uint(userID))
	if err != nil {
		logger.Error("检查用户是否存在失败", "user_id", userID, "error", err)
		resp.FailWithMsg(c, "删除用户失败，请稍后重试")
		return
	}
	if !exists {
		resp.FailWithCode(c, http.StatusNotFound, "用户不存在")
		return
	}

	// 调用服务删除用户
	if err := uc.userModel.DeleteUser(uint(userID)); err != nil {
		logger.Error("删除用户失败", "user_id", userID, "error", err)
		resp.FailWithMsg(c, "删除用户失败")
		return
	}

	resp.OkWithMsg(c, "删除用户成功")
}

// UpdatePassword 更新密码
// @Summary 更新密码
// @Description 用户更新自己的密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body UpdatePasswordRequest true "密码信息"
// @Success 200 {object} resp.Response "更新成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "原密码错误"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/users/password [put]
func (uc *UserController) UpdatePassword(c *gin.Context) {
	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.FailWithValidation(c, err)
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		resp.FailWithCode(c, http.StatusUnauthorized, "未授权")
		return
	}

	// 获取用户信息
	user, err := uc.userModel.GetUserByID(userID.(uint))
	if err != nil {
		logger.Error("获取用户信息失败", "user_id", userID, "error", err)
		resp.FailWithMsg(c, "更新密码失败，请稍后重试")
		return
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		resp.FailWithCode(c, http.StatusForbidden, "原密码错误")
		return
	}

	// 生成新密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("密码加密失败", "error", err)
		resp.FailWithMsg(c, "更新密码失败，请稍后重试")
		return
	}

	// 更新密码
	if err := uc.userModel.UpdatePassword(userID.(uint), string(hashedPassword)); err != nil {
		logger.Error("更新密码失败", "user_id", userID, "error", err)
		resp.FailWithMsg(c, "更新密码失败")
		return
	}

	resp.OkWithMsg(c, "密码更新成功")
}

// GetUserInfo 获取当前用户信息
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} resp.Response "返回用户信息"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/users/info [get]
func (uc *UserController) GetUserInfo(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		resp.FailWithCode(c, http.StatusUnauthorized, "未授权")
		return
	}

	// 获取用户信息
	user, err := uc.userModel.GetUserByID(userID.(uint))
	if err != nil {
		logger.Error("获取用户信息失败", "user_id", userID, "error", err)
		resp.FailWithMsg(c, "获取用户信息失败")
		return
	}

	// 获取用户角色
	roles, err := uc.userModel.GetUserRoles(userID.(uint))
	if err != nil {
		logger.Error("获取用户角色失败", "user_id", userID, "error", err)
		// 这里不返回错误，只是角色可能为空
	}

	// 获取用户权限
	permissions, err := uc.userModel.GetUserPermissions(userID.(uint))
	if err != nil {
		logger.Error("获取用户权限失败", "user_id", userID, "error", err)
		// 这里不返回错误，只是权限可能为空
	}

	// 构建响应数据
	respData := gin.H{
		"user_id":     user.ID,
		"username":    user.Username,
		"email":       user.Email,
		"mobile":      user.Mobile,
		"nickname":    user.Nickname,
		"avatar":      user.Avatar,
		"status":      user.Status,
		"gender":      user.Gender,
		"roles":       roles,
		"permissions": permissions,
	}

	resp.OkWithData(c, respData)
}

// RegisterRoutes 注册路由
func (uc *UserController) RegisterRoutes(router *gin.RouterGroup) {
	userGroup := router.Group("/users")
	{
		// 获取当前用户信息
		userGroup.GET("/info", uc.GetUserInfo)

		// 更新自己的密码
		userGroup.PUT("/password", uc.UpdatePassword)

		// 需要管理员权限的路由
		userGroup.Use(middleware.RequirePermission("system:user:list"))
		{
			// 用户列表
			userGroup.GET("", uc.GetUserList)
		}

		// 用户详情
		userGroup.GET("/:id", middleware.RequirePermission("system:user:info"), uc.GetUserByID)

		// 创建用户
		userGroup.POST("", middleware.RequirePermission("system:user:add"), uc.CreateUser)

		// 更新用户
		userGroup.PUT("/:id", middleware.RequirePermission("system:user:edit"), uc.UpdateUser)

		// 删除用户
		userGroup.DELETE("/:id", middleware.RequirePermission("system:user:delete"), uc.DeleteUser)
	}
}
