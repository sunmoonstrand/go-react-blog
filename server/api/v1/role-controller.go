package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/sunmoonstrand/go-react-blog/server/internal/logger"
	"github.com/sunmoonstrand/go-react-blog/server/internal/middleware"
	"github.com/sunmoonstrand/go-react-blog/server/internal/model"
	"github.com/sunmoonstrand/go-react-blog/server/internal/utils/resp"
)

// RoleController 角色控制器
type RoleController struct {
	roleModel       *model.RoleModel
	permissionModel *model.PermissionModel
}

// NewRoleController 创建角色控制器实例
func NewRoleController(roleModel *model.RoleModel, permissionModel *model.PermissionModel) *RoleController {
	return &RoleController{
		roleModel:       roleModel,
		permissionModel: permissionModel,
	}
}

// CreateRoleRequest 创建角色请求结构
type CreateRoleRequest struct {
	RoleName  string `json:"role_name" binding:"required,min=2,max=50"`
	RoleKey   string `json:"role_key" binding:"required,min=2,max=50"`
	RoleSort  int8   `json:"role_sort" binding:"required,min=0,max=99"`
	RoleDesc  string `json:"role_desc" binding:"max=200"`
	IsDefault bool   `json:"is_default"`
	IsEnabled bool   `json:"is_enabled"`
	PermIDs   []uint `json:"perm_ids" binding:"required,min=1"`
}

// UpdateRoleRequest 更新角色请求结构
type UpdateRoleRequest struct {
	RoleName  string `json:"role_name" binding:"omitempty,min=2,max=50"`
	RoleSort  int8   `json:"role_sort" binding:"omitempty,min=0,max=99"`
	RoleDesc  string `json:"role_desc" binding:"max=200"`
	IsDefault bool   `json:"is_default"`
	IsEnabled bool   `json:"is_enabled"`
	PermIDs   []uint `json:"perm_ids" binding:"omitempty,min=1"`
}

// GetRoleList 获取角色列表
// @Summary 获取角色列表
// @Description 分页获取角色列表
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码，默认1" default(1)
// @Param page_size query int false "每页记录数，默认10" default(10)
// @Param name query string false "角色名称筛选"
// @Param status query int false "状态筛选:0全部,1启用,2禁用"
// @Success 200 {object} resp.Response{data=resp.PageResult} "返回角色列表"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/roles [get]
func (rc *RoleController) GetRoleList(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 获取筛选参数
	name := c.Query("name")
	status, _ := strconv.Atoi(c.DefaultQuery("status", "0"))

	// 调用服务获取角色列表
	var isEnabled *bool
	if status == 1 {
		enabled := true
		isEnabled = &enabled
	} else if status == 2 {
		enabled := false
		isEnabled = &enabled
	}

	roles, total, err := rc.roleModel.GetRoleList(page, pageSize, name, isEnabled)
	if err != nil {
		logger.Error("获取角色列表失败", "error", err)
		resp.FailWithMsg(c, "获取角色列表失败")
		return
	}

	// 返回结果
	resp.OkWithPage(c, roles, total, page, pageSize)
}

// GetRoleByID 根据ID获取角色信息
// @Summary 获取角色详情
// @Description 根据角色ID获取角色详细信息
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "角色ID"
// @Success 200 {object} resp.Response "返回角色详情"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 404 {object} resp.Response "角色不存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/roles/{id} [get]
func (rc *RoleController) GetRoleByID(c *gin.Context) {
	// 获取路径参数
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil {
		resp.FailWithMsg(c, "无效的角色ID")
		return
	}

	// 获取角色信息
	role, err := rc.roleModel.GetRoleByID(uint(roleID))
	if err != nil {
		if err.Error() == "record not found" {
			resp.FailWithCode(c, http.StatusNotFound, "角色不存在")
		} else {
			logger.Error("获取角色信息失败", "role_id", roleID, "error", err)
			resp.FailWithMsg(c, "获取角色信息失败")
		}
		return
	}

	// 获取角色权限
	permissions, err := rc.roleModel.GetRolePermissions(uint(roleID))
	if err != nil {
		logger.Error("获取角色权限失败", "role_id", roleID, "error", err)
		// 这里不返回错误，只是权限可能为空
	}

	// 构建响应数据
	respData := gin.H{
		"role_id":    role.ID,
		"role_name":  role.RoleName,
		"role_key":   role.RoleKey,
		"role_sort":  role.RoleSort,
		"role_desc":  role.RoleDesc,
		"is_default": role.IsDefault,
		"is_enabled": role.IsEnabled,
		"created_at": role.CreatedAt,
		"perms":      permissions,
	}

	resp.OkWithData(c, respData)
}

// CreateRole 创建角色
// @Summary 创建角色
// @Description 创建新角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body CreateRoleRequest true "角色信息"
// @Success 200 {object} resp.Response "创建成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 409 {object} resp.Response "角色名称或标识已存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/roles [post]
func (rc *RoleController) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.FailWithValidation(c, err)
		return
	}

	// 检查角色名称是否已存在
	exists, err := rc.roleModel.CheckRoleNameExists(req.RoleName)
	if err != nil {
		logger.Error("检查角色名称是否存在失败", "role_name", req.RoleName, "error", err)
		resp.FailWithMsg(c, "创建角色失败，请稍后重试")
		return
	}
	if exists {
		resp.FailWithCode(c, http.StatusConflict, "角色名称已存在")
		return
	}

	// 检查角色标识是否已存在
	exists, err = rc.roleModel.CheckRoleKeyExists(req.RoleKey)
	if err != nil {
		logger.Error("检查角色标识是否存在失败", "role_key", req.RoleKey, "error", err)
		resp.FailWithMsg(c, "创建角色失败，请稍后重试")
		return
	}
	if exists {
		resp.FailWithCode(c, http.StatusConflict, "角色标识已存在")
		return
	}

	// 检查权限是否存在
	for _, permID := range req.PermIDs {
		exists, err := rc.permissionModel.CheckPermissionExists(permID)
		if err != nil {
			logger.Error("检查权限是否存在失败", "perm_id", permID, "error", err)
			resp.FailWithMsg(c, "创建角色失败，请稍后重试")
			return
		}
		if !exists {
			resp.FailWithMsg(c, "指定的权限不存在")
			return
		}
	}

	// 创建角色
	role := &model.Role{
		RoleName:  req.RoleName,
		RoleKey:   req.RoleKey,
		RoleSort:  req.RoleSort,
		RoleDesc:  req.RoleDesc,
		IsDefault: req.IsDefault,
		IsEnabled: req.IsEnabled,
	}

	if err := rc.roleModel.CreateRole(role); err != nil {
		logger.Error("创建角色失败", "error", err)
		resp.FailWithMsg(c, "创建角色失败，请稍后重试")
		return
	}

	// 分配权限
	if err := rc.roleModel.AssignPermissions(role.ID, req.PermIDs); err != nil {
		logger.Error("分配权限失败", "role_id", role.ID, "error", err)
		resp.FailWithMsg(c, "角色创建成功，但权限分配失败")
		return
	}

	resp.OkWithMsg(c, "创建角色成功")
}

// UpdateRole 更新角色信息
// @Summary 更新角色信息
// @Description 更新角色信息
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "角色ID"
// @Param data body UpdateRoleRequest true "角色信息"
// @Success 200 {object} resp.Response "更新成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 404 {object} resp.Response "角色不存在"
// @Failure 409 {object} resp.Response "角色名称已被其他角色使用"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/roles/{id} [put]
func (rc *RoleController) UpdateRole(c *gin.Context) {
	// 获取路径参数
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil {
		resp.FailWithMsg(c, "无效的角色ID")
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.FailWithValidation(c, err)
		return
	}

	// 检查角色是否存在
	exists, err := rc.roleModel.CheckRoleExists(uint(roleID))
	if err != nil {
		logger.Error("检查角色是否存在失败", "role_id", roleID, "error", err)
		resp.FailWithMsg(c, "更新角色失败，请稍后重试")
		return
	}
	if !exists {
		resp.FailWithCode(c, http.StatusNotFound, "角色不存在")
		return
	}

	// 如果更新角色名称，检查是否被其他角色使用
	if req.RoleName != "" {
		nameExists, err := rc.roleModel.CheckRoleNameExistsExcept(req.RoleName, uint(roleID))
		if err != nil {
			logger.Error("检查角色名称是否存在失败", "role_name", req.RoleName, "error", err)
			resp.FailWithMsg(c, "更新角色失败，请稍后重试")
			return
		}
		if nameExists {
			resp.FailWithCode(c, http.StatusConflict, "角色名称已被其他角色使用")
			return
		}
	}

	// 如果更新权限，检查权限是否存在
	if len(req.PermIDs) > 0 {
		for _, permID := range req.PermIDs {
			exists, err := rc.permissionModel.CheckPermissionExists(permID)
			if err != nil {
				logger.Error("检查权限是否存在失败", "perm_id", permID, "error", err)
				resp.FailWithMsg(c, "更新角色失败，请稍后重试")
				return
			}
			if !exists {
				resp.FailWithMsg(c, "指定的权限不存在")
				return
			}
		}
	}

	// 构建角色更新信息
	updates := make(map[string]interface{})
	if req.RoleName != "" {
		updates["role_name"] = req.RoleName
	}
	if req.RoleSort != 0 {
		updates["role_sort"] = req.RoleSort
	}
	if req.RoleDesc != "" {
		updates["role_desc"] = req.RoleDesc
	}
	updates["is_default"] = req.IsDefault
	updates["is_enabled"] = req.IsEnabled

	// 更新角色信息
	if len(updates) > 0 {
		if err := rc.roleModel.UpdateRole(uint(roleID), updates); err != nil {
			logger.Error("更新角色信息失败", "role_id", roleID, "error", err)
			resp.FailWithMsg(c, "更新角色信息失败")
			return
		}
	}

	// 更新角色权限
	if len(req.PermIDs) > 0 {
		if err := rc.roleModel.UpdateRolePermissions(uint(roleID), req.PermIDs); err != nil {
			logger.Error("更新角色权限失败", "role_id", roleID, "error", err)
			resp.FailWithMsg(c, "角色信息更新成功，但权限更新失败")
			return
		}
	}

	resp.OkWithMsg(c, "更新角色信息成功")
}

// DeleteRole 删除角色
// @Summary 删除角色
// @Description 根据ID删除角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "角色ID"
// @Success 200 {object} resp.Response "删除成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限或系统内置角色不能删除"
// @Failure 404 {object} resp.Response "角色不存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/roles/{id} [delete]
func (rc *RoleController) DeleteRole(c *gin.Context) {
	// 获取路径参数
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil {
		resp.FailWithMsg(c, "无效的角色ID")
		return
	}

	// 获取角色信息
	role, err := rc.roleModel.GetRoleByID(uint(roleID))
	if err != nil {
		if err.Error() == "record not found" {
			resp.FailWithCode(c, http.StatusNotFound, "角色不存在")
		} else {
			logger.Error("获取角色信息失败", "role_id", roleID, "error", err)
			resp.FailWithMsg(c, "删除角色失败，请稍后重试")
		}
		return
	}

	// 检查是否是内置角色
	if role.RoleKey == "admin" {
		resp.FailWithCode(c, http.StatusForbidden, "超级管理员角色不能删除")
		return
	}

	// 检查角色是否有关联用户
	hasUsers, err := rc.roleModel.CheckRoleHasUsers(uint(roleID))
	if err != nil {
		logger.Error("检查角色是否有关联用户失败", "role_id", roleID, "error", err)
		resp.FailWithMsg(c, "删除角色失败，请稍后重试")
		return
	}
	if hasUsers {
		resp.FailWithMsg(c, "该角色已分配给用户，请先取消用户角色分配")
		return
	}

	// 删除角色
	if err := rc.roleModel.DeleteRole(uint(roleID)); err != nil {
		logger.Error("删除角色失败", "role_id", roleID, "error", err)
		resp.FailWithMsg(c, "删除角色失败")
		return
	}

	resp.OkWithMsg(c, "删除角色成功")
}

// GetAllPermissions 获取所有权限
// @Summary 获取所有权限
// @Description 获取所有权限列表(树形结构)
// @Tags 权限管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} resp.Response "返回权限列表"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/permissions [get]
func (rc *RoleController) GetAllPermissions(c *gin.Context) {
	// 获取所有权限（树形结构）
	permTree, err := rc.permissionModel.GetPermissionTree()
	if err != nil {
		logger.Error("获取权限树失败", "error", err)
		resp.FailWithMsg(c, "获取权限列表失败")
		return
	}

	resp.OkWithData(c, permTree)
}

// GetRolePermissions 获取角色权限
// @Summary 获取角色权限
// @Description 获取指定角色的权限列表
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "角色ID"
// @Success 200 {object} resp.Response "返回角色权限列表"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 404 {object} resp.Response "角色不存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/roles/{id}/permissions [get]
func (rc *RoleController) GetRolePermissions(c *gin.Context) {
	// 获取路径参数
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil {
		resp.FailWithMsg(c, "无效的角色ID")
		return
	}

	// 检查角色是否存在
	exists, err := rc.roleModel.CheckRoleExists(uint(roleID))
	if err != nil {
		logger.Error("检查角色是否存在失败", "role_id", roleID, "error", err)
		resp.FailWithMsg(c, "获取角色权限失败，请稍后重试")
		return
	}
	if !exists {
		resp.FailWithCode(c, http.StatusNotFound, "角色不存在")
		return
	}

	// 获取角色权限
	permissions, err := rc.roleModel.GetRolePermissions(uint(roleID))
	if err != nil {
		logger.Error("获取角色权限失败", "role_id", roleID, "error", err)
		resp.FailWithMsg(c, "获取角色权限失败")
		return
	}

	resp.OkWithData(c, permissions)
}

// RegisterRoutes 注册路由
func (rc *RoleController) RegisterRoutes(router *gin.RouterGroup) {
	roleGroup := router.Group("/roles")
	{
		// 获取角色列表
		roleGroup.GET("", middleware.RequirePermission("system:role:list"), rc.GetRoleList)

		// 创建角色
		roleGroup.POST("", middleware.RequirePermission("system:role:add"), rc.CreateRole)

		// 获取角色详情
		roleGroup.GET("/:id", middleware.RequirePermission("system:role:info"), rc.GetRoleByID)

		// 更新角色
		roleGroup.PUT("/:id", middleware.RequirePermission("system:role:edit"), rc.UpdateRole)

		// 删除角色
		roleGroup.DELETE("/:id", middleware.RequirePermission("system:role:delete"), rc.DeleteRole)

		// 获取角色权限
		roleGroup.GET("/:id/permissions", middleware.RequirePermission("system:role:info"), rc.GetRolePermissions)
	}

	permGroup := router.Group("/permissions")
	{
		// 获取所有权限（树形结构）
		permGroup.GET("", middleware.RequirePermission("system:role:list"), rc.GetAllPermissions)
	}
}
