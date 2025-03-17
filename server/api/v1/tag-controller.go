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

// TagController 标签控制器
type TagController struct {
	tagModel *model.TagModel
}

// NewTagController 创建标签控制器实例
func NewTagController(tagModel *model.TagModel) *TagController {
	return &TagController{
		tagModel: tagModel,
	}
}

// CreateTagRequest 创建标签请求结构
type CreateTagRequest struct {
	TagName     string `json:"tag_name" binding:"required,min=2,max=50"`
	TagKey      string `json:"tag_key" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"max=500"`
	Thumbnail   string `json:"thumbnail" binding:"max=255"`
	SortOrder   int8   `json:"sort_order" binding:"min=0,max=99"`
	IsVisible   bool   `json:"is_visible"`
}

// UpdateTagRequest 更新标签请求结构
type UpdateTagRequest struct {
	TagName     string `json:"tag_name" binding:"omitempty,min=2,max=50"`
	Description string `json:"description" binding:"max=500"`
	Thumbnail   string `json:"thumbnail" binding:"max=255"`
	SortOrder   *int8  `json:"sort_order" binding:"omitempty,min=0,max=99"`
	IsVisible   *bool  `json:"is_visible"`
}

// GetAllTags 获取所有标签
// @Summary 获取所有标签
// @Description 获取所有标签列表
// @Tags 标签管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param include_hidden query bool false "是否包含隐藏标签，默认false"
// @Success 200 {object} resp.Response "返回标签列表"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/tags [get]
func (tc *TagController) GetAllTags(c *gin.Context) {
	// 是否包含隐藏标签
	includeHidden := c.Query("include_hidden") == "true"

	// 调用服务获取标签列表
	tags, err := tc.tagModel.GetAllTags(includeHidden)
	if err != nil {
		logger.Error("获取标签列表失败", "error", err)
		resp.FailWithMsg(c, "获取标签列表失败")
		return
	}

	resp.OkWithData(c, tags)
}

// GetTagList 获取标签列表
// @Summary 获取标签列表
// @Description 分页获取标签列表
// @Tags 标签管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码，默认1" default(1)
// @Param page_size query int false "每页记录数，默认20" default(20)
// @Param name query string false "标签名称筛选"
// @Success 200 {object} resp.Response{data=resp.PageResult} "返回标签列表"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/tags/list [get]
func (tc *TagController) GetTagList(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 获取筛选参数
	name := c.Query("name")

	// 调用服务获取标签列表
	tags, total, err := tc.tagModel.GetTagList(page, pageSize, name)
	if err != nil {
		logger.Error("获取标签列表失败", "error", err)
		resp.FailWithMsg(c, "获取标签列表失败")
		return
	}

	resp.OkWithPage(c, tags, total, page, pageSize)
}

// GetTagByID 根据ID获取标签详情
// @Summary 获取标签详情
// @Description 根据标签ID获取标签详细信息
// @Tags 标签管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "标签ID"
// @Success 200 {object} resp.Response "返回标签详情"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 404 {object} resp.Response "标签不存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/tags/{id} [get]
func (tc *TagController) GetTagByID(c *gin.Context) {
	// 获取路径参数
	tagID, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil {
		resp.FailWithMsg(c, "无效的标签ID")
		return
	}

	// 获取标签详情
	tag, err := tc.tagModel.GetTagByID(uint(tagID))
	if err != nil {
		if err.Error() == "record not found" {
			resp.FailWithCode(c, http.StatusNotFound, "标签不存在")
		} else {
			logger.Error("获取标签详情失败", "tag_id", tagID, "error", err)
			resp.FailWithMsg(c, "获取标签详情失败")
		}
		return
	}

	// 构建响应数据
	respData := gin.H{
		"tag_id":        tag.ID,
		"tag_name":      tag.TagName,
		"tag_key":       tag.TagKey,
		"description":   tag.Description,
		"thumbnail":     tag.Thumbnail,
		"sort_order":    tag.SortOrder,
		"is_visible":    tag.IsVisible,
		"article_count": tag.ArticleCount,
		"created_at":    tag.CreatedAt,
		"updated_at":    tag.UpdatedAt,
	}

	resp.OkWithData(c, respData)
}

// CreateTag 创建标签
// @Summary 创建标签
// @Description 创建新标签
// @Tags 标签管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body CreateTagRequest true "标签信息"
// @Success 200 {object} resp.Response "创建成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 409 {object} resp.Response "标签名称或标识已存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/tags [post]
func (tc *TagController) CreateTag(c *gin.Context) {
	var req CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.FailWithValidation(c, err)
		return
	}

	// 检查标签名称是否已存在
	exists, err := tc.tagModel.CheckTagNameExists(req.TagName)
	if err != nil {
		logger.Error("检查标签名称是否存在失败", "tag_name", req.TagName, "error", err)
		resp.FailWithMsg(c, "创建标签失败，请稍后重试")
		return
	}
	if exists {
		resp.FailWithCode(c, http.StatusConflict, "标签名称已存在")
		return
	}

	// 检查标签标识是否已存在
	exists, err = tc.tagModel.CheckTagKeyExists(req.TagKey)
	if err != nil {
		logger.Error("检查标签标识是否存在失败", "tag_key", req.TagKey, "error", err)
		resp.FailWithMsg(c, "创建标签失败，请稍后重试")
		return
	}
	if exists {
		resp.FailWithCode(c, http.StatusConflict, "标签标识已存在")
		return
	}

	// 创建标签
	tag := &model.Tag{
		TagName:     req.TagName,
		TagKey:      req.TagKey,
		Description: req.Description,
		Thumbnail:   req.Thumbnail,
		SortOrder:   req.SortOrder,
		IsVisible:   req.IsVisible,
	}

	if err := tc.tagModel.CreateTag(tag); err != nil {
		logger.Error("创建标签失败", "error", err)
		resp.FailWithMsg(c, "创建标签失败，请稍后重试")
		return
	}

	resp.OkWithData(c, gin.H{
		"tag_id":  tag.ID,
		"message": "创建标签成功",
	})
}

// UpdateTag 更新标签
// @Summary 更新标签
// @Description 更新标签信息
// @Tags 标签管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "标签ID"
// @Param data body UpdateTagRequest true "标签信息"
// @Success 200 {object} resp.Response "更新成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 404 {object} resp.Response "标签不存在"
// @Failure 409 {object} resp.Response "标签名称已被其他标签使用"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/tags/{id} [put]
func (tc *TagController) UpdateTag(c *gin.Context) {
	// 获取路径参数
	tagID, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil {
		resp.FailWithMsg(c, "无效的标签ID")
		return
	}

	var req UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.FailWithValidation(c, err)
		return
	}

	// 检查标签是否存在
	exists, err := tc.tagModel.CheckTagExists(uint(tagID))
	if err != nil {
		logger.Error("检查标签是否存在失败", "tag_id", tagID, "error", err)
		resp.FailWithMsg(c, "更新标签失败，请稍后重试")
		return
	}
	if !exists {
		resp.FailWithCode(c, http.StatusNotFound, "标签不存在")
		return
	}

	// 如果修改了标签名称，检查是否与其他标签冲突
	if req.TagName != "" {
		exists, err := tc.tagModel.CheckTagNameExistsExcept(req.TagName, uint(tagID))
		if err != nil {
			logger.Error("检查标签名称是否存在失败", "tag_name", req.TagName, "error", err)
			resp.FailWithMsg(c, "更新标签失败，请稍后重试")
			return
		}
		if exists {
			resp.FailWithCode(c, http.StatusConflict, "标签名称已被其他标签使用")
			return
		}
	}

	// 构建标签更新信息
	updates := make(map[string]interface{})
	if req.TagName != "" {
		updates["tag_name"] = req.TagName
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Thumbnail != "" {
		updates["thumbnail"] = req.Thumbnail
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.IsVisible != nil {
		updates["is_visible"] = *req.IsVisible
	}

	// 调用服务更新标签
	if err := tc.tagModel.UpdateTag(uint(tagID), updates); err != nil {
		logger.Error("更新标签失败", "tag_id", tagID, "error", err)
		resp.FailWithMsg(c, "更新标签失败，请稍后重试")
		return
	}

	resp.OkWithMsg(c, "更新标签成功")
}

// DeleteTag 删除标签
// @Summary 删除标签
// @Description 根据ID删除标签
// @Tags 标签管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "标签ID"
// @Success 200 {object} resp.Response "删除成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 404 {object} resp.Response "标签不存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/tags/{id} [delete]
func (tc *TagController) DeleteTag(c *gin.Context) {
	// 获取路径参数
	tagID, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil {
		resp.FailWithMsg(c, "无效的标签ID")
		return
	}

	// 检查标签是否存在
	exists, err := tc.tagModel.CheckTagExists(uint(tagID))
	if err != nil {
		logger.Error("检查标签是否存在失败", "tag_id", tagID, "error", err)
		resp.FailWithMsg(c, "删除标签失败，请稍后重试")
		return
	}
	if !exists {
		resp.FailWithCode(c, http.StatusNotFound, "标签不存在")
		return
	}

	// 检查标签是否有关联文章
	hasArticles, err := tc.tagModel.HasArticles(uint(tagID))
	if err != nil {
		logger.Error("检查是否有关联文章失败", "tag_id", tagID, "error", err)
		resp.FailWithMsg(c, "删除标签失败，请稍后重试")
		return
	}
	if hasArticles {
		resp.FailWithMsg(c, "该标签下有关联文章，不能删除")
		return
	}

	// 调用服务删除标签
	if err := tc.tagModel.DeleteTag(uint(tagID)); err != nil {
		logger.Error("删除标签失败", "tag_id", tagID, "error", err)
		resp.FailWithMsg(c, "删除标签失败")
		return
	}

	resp.OkWithMsg(c, "删除标签成功")
}

// BatchDeleteTags 批量删除标签
// @Summary 批量删除标签
// @Description 批量删除多个标签
// @Tags 标签管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param ids body []int true "标签ID数组"
// @Success 200 {object} resp.Response "删除成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/tags/batch [delete]
func (tc *TagController) BatchDeleteTags(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		resp.FailWithMsg(c, "无效的参数")
		return
	}

	if len(ids) == 0 {
		resp.FailWithMsg(c, "请选择要删除的标签")
		return
	}

	// 调用服务批量删除标签
	result, err := tc.tagModel.BatchDeleteTags(ids)
	if err != nil {
		logger.Error("批量删除标签失败", "error", err)
		resp.FailWithMsg(c, "批量删除标签失败")
		return
	}

	resp.OkWithData(c, gin.H{
		"success": result.Success,
		"failed":  result.Failed,
		"message": "批量删除标签完成",
	})
}

// RegisterRoutes 注册路由
func (tc *TagController) RegisterRoutes(router *gin.RouterGroup) {
	tagGroup := router.Group("/tags")
	{
		// 获取所有标签
		tagGroup.GET("", tc.GetAllTags)

		// 获取标签列表(分页)
		tagGroup.GET("/list", middleware.RequirePermission("content:tag:list"), tc.GetTagList)

		// 获取标签详情
		tagGroup.GET("/:id", middleware.RequirePermission("content:tag:info"), tc.GetTagByID)

		// 需要管理员权限的路由
		// 创建标签
		tagGroup.POST("", middleware.RequirePermission("content:tag:add"), tc.CreateTag)

		// 更新标签
		tagGroup.PUT("/:id", middleware.RequirePermission("content:tag:edit"), tc.UpdateTag)

		// 删除标签
		tagGroup.DELETE("/:id", middleware.RequirePermission("content:tag:delete"), tc.DeleteTag)

		// 批量删除标签
		tagGroup.DELETE("/batch", middleware.RequirePermission("content:tag:delete"), tc.BatchDeleteTags)
	}
}
