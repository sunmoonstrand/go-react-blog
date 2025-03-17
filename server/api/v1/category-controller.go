package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/yourusername/blog/server/internal/logger"
	"github.com/yourusername/blog/server/internal/middleware"
	"github.com/yourusername/blog/server/internal/model"
	"github.com/yourusername/blog/server/internal/utils/resp"
)

// CategoryController 分类控制器
type CategoryController struct {
	categoryModel *model.CategoryModel
}

// NewCategoryController 创建分类控制器实例
func NewCategoryController(categoryModel *model.CategoryModel) *CategoryController {
	return &CategoryController{
		categoryModel: categoryModel,
	}
}

// CreateCategoryRequest 创建分类请求结构
type CreateCategoryRequest struct {
	ParentID       uint   `json:"parent_id"`
	CategoryName   string `json:"category_name" binding:"required,min=2,max=50"`
	CategoryKey    string `json:"category_key" binding:"required,min=2,max=50"`
	Description    string `json:"description" binding:"max=500"`
	Thumbnail      string `json:"thumbnail" binding:"max=255"`
	Icon           string `json:"icon" binding:"max=100"`
	SortOrder      int8   `json:"sort_order" binding:"min=0,max=99"`
	IsVisible      bool   `json:"is_visible"`
	SEOTitle       string `json:"seo_title" binding:"max=100"`
	SEOKeywords    string `json:"seo_keywords" binding:"max=200"`
	SEODescription string `json:"seo_description" binding:"max=300"`
}

// UpdateCategoryRequest 更新分类请求结构
type UpdateCategoryRequest struct {
	ParentID       uint   `json:"parent_id"`
	CategoryName   string `json:"category_name" binding:"omitempty,min=2,max=50"`
	Description    string `json:"description" binding:"max=500"`
	Thumbnail      string `json:"thumbnail" binding:"max=255"`
	Icon           string `json:"icon" binding:"max=100"`
	SortOrder      *int8  `json:"sort_order" binding:"omitempty,min=0,max=99"`
	IsVisible      *bool  `json:"is_visible"`
	SEOTitle       string `json:"seo_title" binding:"max=100"`
	SEOKeywords    string `json:"seo_keywords" binding:"max=200"`
	SEODescription string `json:"seo_description" binding:"max=300"`
}

// GetAllCategories 获取所有分类
// @Summary 获取所有分类
// @Description 获取所有分类(树形结构)
// @Tags 分类管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param include_hidden query bool false "是否包含隐藏分类，默认false"
// @Success 200 {object} resp.Response "返回分类列表"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/categories [get]
func (cc *CategoryController) GetAllCategories(c *gin.Context) {
	// 是否包含隐藏分类
	includeHidden := c.Query("include_hidden") == "true"

	// 调用服务获取分类列表
	categories, err := cc.categoryModel.GetCategoryTree(includeHidden)
	if err != nil {
		logger.Error("获取分类列表失败", "error", err)
		resp.FailWithMsg(c, "获取分类列表失败")
		return
	}

	resp.OkWithData(c, categories)
}

// GetCategoryList 获取分类列表
// @Summary 获取分类列表
// @Description 分页获取分类列表(非树形)
// @Tags 分类管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码，默认1" default(1)
// @Param page_size query int false "每页记录数，默认20" default(20)
// @Param name query string false "分类名称筛选"
// @Param parent_id query int false "父分类ID筛选"
// @Success 200 {object} resp.Response{data=resp.PageResult} "返回分类列表"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/categories/list [get]
func (cc *CategoryController) GetCategoryList(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 获取筛选参数
	name := c.Query("name")
	parentID, _ := strconv.Atoi(c.DefaultQuery("parent_id", "-1"))

	// 调用服务获取分类列表
	categories, total, err := cc.categoryModel.GetCategoryList(page, pageSize, name, parentID)
	if err != nil {
		logger.Error("获取分类列表失败", "error", err)
		resp.FailWithMsg(c, "获取分类列表失败")
		return
	}

	resp.OkWithPage(c, categories, total, page, pageSize)
}

// GetCategoryByID 根据ID获取分类详情
// @Summary 获取分类详情
// @Description 根据分类ID获取分类详细信息
// @Tags 分类管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "分类ID"
// @Success 200 {object} resp.Response "返回分类详情"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 404 {object} resp.Response "分类不存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/categories/{id} [get]
func (cc *CategoryController) GetCategoryByID(c *gin.Context) {
	// 获取路径参数
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil {
		resp.FailWithMsg(c, "无效的分类ID")
		return
	}

	// 获取分类详情
	category, err := cc.categoryModel.GetCategoryByID(uint(categoryID))
	if err != nil {
		if err.Error() == "record not found" {
			resp.FailWithCode(c, http.StatusNotFound, "分类不存在")
		} else {
			logger.Error("获取分类详情失败", "category_id", categoryID, "error", err)
			resp.FailWithMsg(c, "获取分类详情失败")
		}
		return
	}

	// 获取子分类
	children, err := cc.categoryModel.GetChildCategories(uint(categoryID))
	if err != nil {
		logger.Error("获取子分类失败", "parent_id", categoryID, "error", err)
		// 这里不返回错误，只是子分类可能为空
	}

	// 构建响应数据
	respData := gin.H{
		"category_id":     category.ID,
		"parent_id":       category.ParentID,
		"category_name":   category.CategoryName,
		"category_key":    category.CategoryKey,
		"path":            category.Path,
		"description":     category.Description,
		"thumbnail":       category.Thumbnail,
		"icon":            category.Icon,
		"sort_order":      category.SortOrder,
		"is_visible":      category.IsVisible,
		"seo_title":       category.SEOTitle,
		"seo_keywords":    category.SEOKeywords,
		"seo_description": category.SEODescription,
		"article_count":   category.ArticleCount,
		"created_at":      category.CreatedAt,
		"updated_at":      category.UpdatedAt,
		"children":        children,
	}

	resp.OkWithData(c, respData)
}

// CreateCategory 创建分类
// @Summary 创建分类
// @Description 创建新分类
// @Tags 分类管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body CreateCategoryRequest true "分类信息"
// @Success 200 {object} resp.Response "创建成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 409 {object} resp.Response "分类名称或标识已存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/categories [post]
func (cc *CategoryController) CreateCategory(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.FailWithValidation(c, err)
		return
	}

	// 检查分类名称是否已存在
	exists, err := cc.categoryModel.CheckCategoryNameExists(req.CategoryName)
	if err != nil {
		logger.Error("检查分类名称是否存在失败", "category_name", req.CategoryName, "error", err)
		resp.FailWithMsg(c, "创建分类失败，请稍后重试")
		return
	}
	if exists {
		resp.FailWithCode(c, http.StatusConflict, "分类名称已存在")
		return
	}

	// 检查分类标识是否已存在
	exists, err = cc.categoryModel.CheckCategoryKeyExists(req.CategoryKey)
	if err != nil {
		logger.Error("检查分类标识是否存在失败", "category_key", req.CategoryKey, "error", err)
		resp.FailWithMsg(c, "创建分类失败，请稍后重试")
		return
	}
	if exists {
		resp.FailWithCode(c, http.StatusConflict, "分类标识已存在")
		return
	}

	// 如果有父分类，检查父分类是否存在
	if req.ParentID > 0 {
		exists, err = cc.categoryModel.CheckCategoryExists(req.ParentID)
		if err != nil {
			logger.Error("检查父分类是否存在失败", "parent_id", req.ParentID, "error", err)
			resp.FailWithMsg(c, "创建分类失败，请稍后重试")
			return
		}
		if !exists {
			resp.FailWithMsg(c, "父分类不存在")
			return
		}
	}

	// 创建分类
	category := &model.Category{
		ParentID:       req.ParentID,
		CategoryName:   req.CategoryName,
		CategoryKey:    req.CategoryKey,
		Description:    req.Description,
		Thumbnail:      req.Thumbnail,
		Icon:           req.Icon,
		SortOrder:      req.SortOrder,
		IsVisible:      req.IsVisible,
		SEOTitle:       req.SEOTitle,
		SEOKeywords:    req.SEOKeywords,
		SEODescription: req.SEODescription,
	}

	if err := cc.categoryModel.CreateCategory(category); err != nil {
		logger.Error("创建分类失败", "error", err)
		resp.FailWithMsg(c, "创建分类失败，请稍后重试")
		return
	}

	resp.OkWithData(c, gin.H{
		"category_id": category.ID,
		"message":     "创建分类成功",
	})
}

// UpdateCategory 更新分类
// @Summary 更新分类
// @Description 更新分类信息
// @Tags 分类管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "分类ID"
// @Param data body UpdateCategoryRequest true "分类信息"
// @Success 200 {object} resp.Response "更新成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 404 {object} resp.Response "分类不存在"
// @Failure 409 {object} resp.Response "分类名称已被其他分类使用"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/categories/{id} [put]
func (cc *CategoryController) UpdateCategory(c *gin.Context) {
	// 获取路径参数
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil {
		resp.FailWithMsg(c, "无效的分类ID")
		return
	}

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.FailWithValidation(c, err)
		return
	}

	// 检查分类是否存在
	exists, err := cc.categoryModel.CheckCategoryExists(uint(categoryID))
	if err != nil {
		logger.Error("检查分类是否存在失败", "category_id", categoryID, "error", err)
		resp.FailWithMsg(c, "更新分类失败，请稍后重试")
		return
	}
	if !exists {
		resp.FailWithCode(c, http.StatusNotFound, "分类不存在")
		return
	}

	// 如果修改了分类名称，检查是否与其他分类冲突
	if req.CategoryName != "" {
		exists, err := cc.categoryModel.CheckCategoryNameExistsExcept(req.CategoryName, uint(categoryID))
		if err != nil {
			logger.Error("检查分类名称是否存在失败", "category_name", req.CategoryName, "error", err)
			resp.FailWithMsg(c, "更新分类失败，请稍后重试")
			return
		}
		if exists {
			resp.FailWithCode(c, http.StatusConflict, "分类名称已被其他分类使用")
			return
		}
	}

	// 父分类不能是自己或自己的子分类
	if req.ParentID > 0 && req.ParentID != 0 {
		// 不能设置自己为父分类
		if req.ParentID == uint(categoryID) {
			resp.FailWithMsg(c, "不能将自己设为父分类")
			return
		}

		// 检查父分类是否存在
		exists, err = cc.categoryModel.CheckCategoryExists(req.ParentID)
		if err != nil {
			logger.Error("检查父分类是否存在失败", "parent_id", req.ParentID, "error", err)
			resp.FailWithMsg(c, "更新分类失败，请稍后重试")
			return
		}
		if !exists {
			resp.FailWithMsg(c, "父分类不存在")
			return
		}

		// 检查是否会形成循环依赖
		isChild, err := cc.categoryModel.IsChildCategory(uint(categoryID), req.ParentID)
		if err != nil {
			logger.Error("检查是否为子分类失败", "category_id", categoryID, "parent_id", req.ParentID, "error", err)
			resp.FailWithMsg(c, "更新分类失败，请稍后重试")
			return
		}
		if isChild {
			resp.FailWithMsg(c, "不能将子分类设为父分类")
			return
		}
	}

	// 构建分类更新信息
	updates := make(map[string]interface{})
	updates["parent_id"] = req.ParentID // 即使是0也要更新，表示设为根分类
	if req.CategoryName != "" {
		updates["category_name"] = req.CategoryName
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Thumbnail != "" {
		updates["thumbnail"] = req.Thumbnail
	}
	if req.Icon != "" {
		updates["icon"] = req.Icon
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.IsVisible != nil {
		updates["is_visible"] = *req.IsVisible
	}
	if req.SEOTitle != "" {
		updates["seo_title"] = req.SEOTitle
	}
	if req.SEOKeywords != "" {
		updates["seo_keywords"] = req.SEOKeywords
	}
	if req.SEODescription != "" {
		updates["seo_description"] = req.SEODescription
	}

	// 调用服务更新分类
	if err := cc.categoryModel.UpdateCategory(uint(categoryID), updates); err != nil {
		logger.Error("更新分类失败", "category_id", categoryID, "error", err)
		resp.FailWithMsg(c, "更新分类失败，请稍后重试")
		return
	}

	resp.OkWithMsg(c, "更新分类成功")
}

// DeleteCategory 删除分类
// @Summary 删除分类
// @Description 根据ID删除分类
// @Tags 分类管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "分类ID"
// @Success 200 {object} resp.Response "删除成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 404 {object} resp.Response "分类不存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/categories/{id} [delete]
func (cc *CategoryController) DeleteCategory(c *gin.Context) {
	// 获取路径参数
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil {
		resp.FailWithMsg(c, "无效的分类ID")
		return
	}

	// 检查分类是否存在
	exists, err := cc.categoryModel.CheckCategoryExists(uint(categoryID))
	if err != nil {
		logger.Error("检查分类是否存在失败", "category_id", categoryID, "error", err)
		resp.FailWithMsg(c, "删除分类失败，请稍后重试")
		return
	}
	if !exists {
		resp.FailWithCode(c, http.StatusNotFound, "分类不存在")
		return
	}

	// 检查分类是否有子分类
	hasChildren, err := cc.categoryModel.HasChildCategories(uint(categoryID))
	if err != nil {
		logger.Error("检查是否有子分类失败", "category_id", categoryID, "error", err)
		resp.FailWithMsg(c, "删除分类失败，请稍后重试")
		return
	}
	if hasChildren {
		resp.FailWithMsg(c, "该分类下有子分类，不能删除")
		return
	}

	// 检查分类是否有关联文章
	hasArticles, err := cc.categoryModel.HasArticles(uint(categoryID))
	if err != nil {
		logger.Error("检查是否有关联文章失败", "category_id", categoryID, "error", err)
		resp.FailWithMsg(c, "删除分类失败，请稍后重试")
		return
	}
	if hasArticles {
		resp.FailWithMsg(c, "该分类下有关联文章，不能删除")
		return
	}

	// 调用服务删除分类
	if err := cc.categoryModel.DeleteCategory(uint(categoryID)); err != nil {
		logger.Error("删除分类失败", "category_id", categoryID, "error", err)
		resp.FailWithMsg(c, "删除分类失败")
		return
	}

	resp.OkWithMsg(c, "删除分类成功")
}

// BatchDeleteCategories 批量删除分类
// @Summary 批量删除分类
// @Description 批量删除多个分类
// @Tags 分类管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param ids body []int true "分类ID数组"
// @Success 200 {object} resp.Response "删除成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/categories/batch [delete]
func (cc *CategoryController) BatchDeleteCategories(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		resp.FailWithMsg(c, "无效的参数")
		return
	}

	if len(ids) == 0 {
		resp.FailWithMsg(c, "请选择要删除的分类")
		return
	}

	// 调用服务批量删除分类
	result, err := cc.categoryModel.BatchDeleteCategories(ids)
	if err != nil {
		logger.Error("批量删除分类失败", "error", err)
		resp.FailWithMsg(c, "批量删除分类失败")
		return
	}

	resp.OkWithData(c, gin.H{
		"success": result.Success,
		"failed":  result.Failed,
		"message": "批量删除分类完成",
	})
}

// RegisterRoutes 注册路由
func (cc *CategoryController) RegisterRoutes(router *gin.RouterGroup) {
	categoryGroup := router.Group("/categories")
	{
		// 获取所有分类(树形结构)
		categoryGroup.GET("", cc.GetAllCategories)

		// 获取分类列表(分页，非树形)
		categoryGroup.GET("/list", middleware.RequirePermission("content:category:list"), cc.GetCategoryList)

		// 获取分类详情
		categoryGroup.GET("/:id", middleware.RequirePermission("content:category:info"), cc.GetCategoryByID)

		// 需要管理员权限的路由
		// 创建分类
		categoryGroup.POST("", middleware.RequirePermission("content:category:add"), cc.CreateCategory)

		// 更新分类
		categoryGroup.PUT("/:id", middleware.RequirePermission("content:category:edit"), cc.UpdateCategory)

		// 删除分类
		categoryGroup.DELETE("/:id", middleware.RequirePermission("content:category:delete"), cc.DeleteCategory)

		// 批量删除分类
		categoryGroup.DELETE("/batch", middleware.RequirePermission("content:category:delete"), cc.BatchDeleteCategories)
	}
}
