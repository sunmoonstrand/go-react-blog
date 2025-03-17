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

// ArticleController 文章控制器
type ArticleController struct {
	articleModel  *model.ArticleModel
	categoryModel *model.CategoryModel
	tagModel      *model.TagModel
}

// NewArticleController 创建文章控制器实例
func NewArticleController(articleModel *model.ArticleModel, categoryModel *model.CategoryModel, tagModel *model.TagModel) *ArticleController {
	return &ArticleController{
		articleModel:  articleModel,
		categoryModel: categoryModel,
		tagModel:      tagModel,
	}
}

// CreateArticleRequest 创建文章请求结构
type CreateArticleRequest struct {
	Title         string `json:"title" binding:"required,min=2,max=200"`
	ArticleKey    string `json:"article_key" binding:"required,min=2,max=200"`
	Summary       string `json:"summary" binding:"max=500"`
	Thumbnail     string `json:"thumbnail"`
	Content       string `json:"content" binding:"required"`
	ContentFormat uint8  `json:"content_format" binding:"required,oneof=1 2"`
	Status        uint8  `json:"status" binding:"required,oneof=1 2 3 4"`
	ArticleType   uint8  `json:"article_type" binding:"required,oneof=1 2 3 4"`
	CategoryIDs   []uint `json:"category_ids" binding:"required,min=1"`
	TagIDs        []uint `json:"tag_ids"`
	IsTop         bool   `json:"is_top"`
	IsRecommend   bool   `json:"is_recommend"`
	AllowComment  bool   `json:"allow_comment"`
	SEOTitle      string `json:"seo_title" binding:"max=100"`
	SEOKeywords   string `json:"seo_keywords" binding:"max=200"`
	SEODesc       string `json:"seo_description" binding:"max=300"`
	SourceURL     string `json:"source_url" binding:"max=255"`
	SourceName    string `json:"source_name" binding:"max=100"`
}

// UpdateArticleRequest 更新文章请求结构
type UpdateArticleRequest struct {
	Title         string `json:"title" binding:"omitempty,min=2,max=200"`
	Summary       string `json:"summary" binding:"max=500"`
	Thumbnail     string `json:"thumbnail"`
	Content       string `json:"content"`
	ContentFormat uint8  `json:"content_format" binding:"omitempty,oneof=1 2"`
	Status        uint8  `json:"status" binding:"omitempty,oneof=1 2 3 4"`
	ArticleType   uint8  `json:"article_type" binding:"omitempty,oneof=1 2 3 4"`
	CategoryIDs   []uint `json:"category_ids" binding:"omitempty,min=1"`
	TagIDs        []uint `json:"tag_ids"`
	IsTop         *bool  `json:"is_top"`
	IsRecommend   *bool  `json:"is_recommend"`
	AllowComment  *bool  `json:"allow_comment"`
	SEOTitle      string `json:"seo_title" binding:"max=100"`
	SEOKeywords   string `json:"seo_keywords" binding:"max=200"`
	SEODesc       string `json:"seo_description" binding:"max=300"`
	SourceURL     string `json:"source_url" binding:"max=255"`
	SourceName    string `json:"source_name" binding:"max=100"`
}

// GetArticleList 获取文章列表
// @Summary 获取文章列表
// @Description 分页获取文章列表
// @Tags 文章管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码，默认1" default(1)
// @Param page_size query int false "每页记录数，默认10" default(10)
// @Param keyword query string false "关键词搜索(标题)"
// @Param status query int false "状态筛选:0全部,1草稿,2待审核,3已发布,4已下线"
// @Param category_id query int false "分类ID筛选"
// @Param tag_id query int false "标签ID筛选"
// @Param type query int false "文章类型:0全部,1原创,2转载,3翻译"
// @Success 200 {object} resp.Response{data=resp.PageResult} "返回文章列表"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/articles [get]
func (ac *ArticleController) GetArticleList(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 获取筛选参数
	keyword := c.Query("keyword")
	status, _ := strconv.Atoi(c.DefaultQuery("status", "0"))
	categoryID, _ := strconv.Atoi(c.DefaultQuery("category_id", "0"))
	tagID, _ := strconv.Atoi(c.DefaultQuery("tag_id", "0"))
	articleType, _ := strconv.Atoi(c.DefaultQuery("type", "0"))

	// 调用服务获取文章列表
	articles, total, err := ac.articleModel.GetArticleList(page, pageSize, keyword, uint8(status), uint(categoryID), uint(tagID), uint8(articleType))
	if err != nil {
		logger.Error("获取文章列表失败", "error", err)
		resp.FailWithMsg(c, "获取文章列表失败")
		return
	}

	// 返回结果
	resp.OkWithPage(c, articles, total, page, pageSize)
}

// GetArticleByID 根据ID获取文章详情
// @Summary 获取文章详情
// @Description 根据文章ID获取文章详细信息
// @Tags 文章管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "文章ID"
// @Success 200 {object} resp.Response "返回文章详情"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 404 {object} resp.Response "文章不存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/articles/{id} [get]
func (ac *ArticleController) GetArticleByID(c *gin.Context) {
	// 获取路径参数
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil {
		resp.FailWithMsg(c, "无效的文章ID")
		return
	}

	// 获取文章详情
	article, err := ac.articleModel.GetArticleByID(uint(articleID))
	if err != nil {
		if err.Error() == "record not found" {
			resp.FailWithCode(c, http.StatusNotFound, "文章不存在")
		} else {
			logger.Error("获取文章详情失败", "article_id", articleID, "error", err)
			resp.FailWithMsg(c, "获取文章详情失败")
		}
		return
	}

	// 获取文章内容
	content, err := ac.articleModel.GetArticleContent(uint(articleID))
	if err != nil {
		logger.Error("获取文章内容失败", "article_id", articleID, "error", err)
		resp.FailWithMsg(c, "获取文章内容失败")
		return
	}

	// 获取文章分类
	categories, err := ac.articleModel.GetArticleCategories(uint(articleID))
	if err != nil {
		logger.Error("获取文章分类失败", "article_id", articleID, "error", err)
		// 这里不返回错误，只是分类可能为空
	}

	// 获取文章标签
	tags, err := ac.articleModel.GetArticleTags(uint(articleID))
	if err != nil {
		logger.Error("获取文章标签失败", "article_id", articleID, "error", err)
		// 这里不返回错误，只是标签可能为空
	}

	// 构建响应数据
	respData := gin.H{
		"article_id":      article.ID,
		"title":           article.Title,
		"article_key":     article.ArticleKey,
		"summary":         article.Summary,
		"thumbnail":       article.Thumbnail,
		"status":          article.Status,
		"article_type":    article.ArticleType,
		"view_count":      article.ViewCount,
		"like_count":      article.LikeCount,
		"comment_count":   article.CommentCount,
		"allow_comment":   article.AllowComment,
		"is_top":          article.IsTop,
		"is_recommend":    article.IsRecommend,
		"seo_title":       article.SEOTitle,
		"seo_keywords":    article.SEOKeywords,
		"seo_description": article.SEODescription,
		"source_url":      article.SourceURL,
		"source_name":     article.SourceName,
		"content":         content.Content,
		"content_format":  content.ContentFormat,
		"publish_time":    article.PublishTime,
		"created_at":      article.CreatedAt,
		"updated_at":      article.UpdatedAt,
		"categories":      categories,
		"tags":            tags,
	}

	resp.OkWithData(c, respData)
}

// CreateArticle 创建文章
// @Summary 创建文章
// @Description 创建新文章
// @Tags 文章管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body CreateArticleRequest true "文章信息"
// @Success 200 {object} resp.Response "创建成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 409 {object} resp.Response "文章标识已存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/articles [post]
func (ac *ArticleController) CreateArticle(c *gin.Context) {
	var req CreateArticleRequest
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

	// 检查文章标识是否已存在
	exists, err := ac.articleModel.CheckArticleKeyExists(req.ArticleKey)
	if err != nil {
		logger.Error("检查文章标识是否存在失败", "article_key", req.ArticleKey, "error", err)
		resp.FailWithMsg(c, "创建文章失败，请稍后重试")
		return
	}
	if exists {
		resp.FailWithCode(c, http.StatusConflict, "文章标识已存在")
		return
	}

	// 检查分类是否存在
	for _, categoryID := range req.CategoryIDs {
		exists, err := ac.categoryModel.CheckCategoryExists(categoryID)
		if err != nil {
			logger.Error("检查分类是否存在失败", "category_id", categoryID, "error", err)
			resp.FailWithMsg(c, "创建文章失败，请稍后重试")
			return
		}
		if !exists {
			resp.FailWithMsg(c, "指定的分类不存在")
			return
		}
	}

	// 检查标签是否存在
	if len(req.TagIDs) > 0 {
		for _, tagID := range req.TagIDs {
			exists, err := ac.tagModel.CheckTagExists(tagID)
			if err != nil {
				logger.Error("检查标签是否存在失败", "tag_id", tagID, "error", err)
				resp.FailWithMsg(c, "创建文章失败，请稍后重试")
				return
			}
			if !exists {
				resp.FailWithMsg(c, "指定的标签不存在")
				return
			}
		}
	}

	// 创建文章
	article := &model.Article{
		UserID:         userID.(uint),
		Title:          req.Title,
		ArticleKey:     req.ArticleKey,
		Summary:        req.Summary,
		Thumbnail:      req.Thumbnail,
		Status:         req.Status,
		ArticleType:    req.ArticleType,
		AllowComment:   req.AllowComment,
		IsTop:          req.IsTop,
		IsRecommend:    req.IsRecommend,
		SEOTitle:       req.SEOTitle,
		SEOKeywords:    req.SEOKeywords,
		SEODescription: req.SEODesc,
		SourceURL:      req.SourceURL,
		SourceName:     req.SourceName,
	}

	// 创建文章内容
	content := &model.ArticleContent{
		Content:       req.Content,
		ContentFormat: req.ContentFormat,
		Version:       1,
		IsCurrent:     true,
	}

	// 调用服务创建文章
	articleID, err := ac.articleModel.CreateArticle(article, content, req.CategoryIDs, req.TagIDs)
	if err != nil {
		logger.Error("创建文章失败", "error", err)
		resp.FailWithMsg(c, "创建文章失败，请稍后重试")
		return
	}

	resp.OkWithData(c, gin.H{
		"article_id": articleID,
		"message":    "创建文章成功",
	})
}

// UpdateArticle 更新文章
// @Summary 更新文章
// @Description 更新文章信息
// @Tags 文章管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "文章ID"
// @Param data body UpdateArticleRequest true "文章信息"
// @Success 200 {object} resp.Response "更新成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 404 {object} resp.Response "文章不存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/articles/{id} [put]
func (ac *ArticleController) UpdateArticle(c *gin.Context) {
	// 获取路径参数
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil {
		resp.FailWithMsg(c, "无效的文章ID")
		return
	}

	var req UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.FailWithValidation(c, err)
		return
	}

	// 检查文章是否存在
	exists, err := ac.articleModel.CheckArticleExists(uint(articleID))
	if err != nil {
		logger.Error("检查文章是否存在失败", "article_id", articleID, "error", err)
		resp.FailWithMsg(c, "更新文章失败，请稍后重试")
		return
	}
	if !exists {
		resp.FailWithCode(c, http.StatusNotFound, "文章不存在")
		return
	}

	// 获取当前用户ID和角色
	userID, exists := c.Get("user_id")
	if !exists {
		resp.FailWithCode(c, http.StatusUnauthorized, "未授权")
		return
	}

	// 检查权限 - 只有文章作者或管理员可以修改文章
	isAdmin, _ := c.Get("is_admin")
	isAuthor, err := ac.articleModel.CheckIsAuthor(uint(articleID), userID.(uint))
	if err != nil {
		logger.Error("检查文章作者失败", "article_id", articleID, "user_id", userID, "error", err)
		resp.FailWithMsg(c, "更新文章失败，请稍后重试")
		return
	}

	if !isAdmin.(bool) && !isAuthor {
		resp.FailWithCode(c, http.StatusForbidden, "只有文章作者或管理员可以修改文章")
		return
	}

	// 检查分类是否存在
	if len(req.CategoryIDs) > 0 {
		for _, categoryID := range req.CategoryIDs {
			exists, err := ac.categoryModel.CheckCategoryExists(categoryID)
			if err != nil {
				logger.Error("检查分类是否存在失败", "category_id", categoryID, "error", err)
				resp.FailWithMsg(c, "更新文章失败，请稍后重试")
				return
			}
			if !exists {
				resp.FailWithMsg(c, "指定的分类不存在")
				return
			}
		}
	}

	// 检查标签是否存在
	if len(req.TagIDs) > 0 {
		for _, tagID := range req.TagIDs {
			exists, err := ac.tagModel.CheckTagExists(tagID)
			if err != nil {
				logger.Error("检查标签是否存在失败", "tag_id", tagID, "error", err)
				resp.FailWithMsg(c, "更新文章失败，请稍后重试")
				return
			}
			if !exists {
				resp.FailWithMsg(c, "指定的标签不存在")
				return
			}
		}
	}

	// 构建文章更新信息
	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Summary != "" {
		updates["summary"] = req.Summary
	}
	if req.Thumbnail != "" {
		updates["thumbnail"] = req.Thumbnail
	}
	if req.Status != 0 {
		updates["status"] = req.Status
	}
	if req.ArticleType != 0 {
		updates["article_type"] = req.ArticleType
	}
	if req.IsTop != nil {
		updates["is_top"] = *req.IsTop
	}
	if req.IsRecommend != nil {
		updates["is_recommend"] = *req.IsRecommend
	}
	if req.AllowComment != nil {
		updates["allow_comment"] = *req.AllowComment
	}
	if req.SEOTitle != "" {
		updates["seo_title"] = req.SEOTitle
	}
	if req.SEOKeywords != "" {
		updates["seo_keywords"] = req.SEOKeywords
	}
	if req.SEODesc != "" {
		updates["seo_description"] = req.SEODesc
	}
	if req.SourceURL != "" {
		updates["source_url"] = req.SourceURL
	}
	if req.SourceName != "" {
		updates["source_name"] = req.SourceName
	}

	// 创建新的文章内容版本（如果内容有更新）
	var newContent *model.ArticleContent
	if req.Content != "" {
		newContent = &model.ArticleContent{
			ArticleID:     uint(articleID),
			Content:       req.Content,
			ContentFormat: req.ContentFormat,
			IsCurrent:     true,
		}
	}

	// 调用服务更新文章
	err = ac.articleModel.UpdateArticle(uint(articleID), updates, newContent, req.CategoryIDs, req.TagIDs)
	if err != nil {
		logger.Error("更新文章失败", "article_id", articleID, "error", err)
		resp.FailWithMsg(c, "更新文章失败，请稍后重试")
		return
	}

	resp.OkWithMsg(c, "更新文章成功")
}

// DeleteArticle 删除文章
// @Summary 删除文章
// @Description 根据ID删除文章
// @Tags 文章管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "文章ID"
// @Success 200 {object} resp.Response "删除成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 404 {object} resp.Response "文章不存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/articles/{id} [delete]
func (ac *ArticleController) DeleteArticle(c *gin.Context) {
	// 获取路径参数
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil {
		resp.FailWithMsg(c, "无效的文章ID")
		return
	}

	// 检查文章是否存在
	exists, err := ac.articleModel.CheckArticleExists(uint(articleID))
	if err != nil {
		logger.Error("检查文章是否存在失败", "article_id", articleID, "error", err)
		resp.FailWithMsg(c, "删除文章失败，请稍后重试")
		return
	}
	if !exists {
		resp.FailWithCode(c, http.StatusNotFound, "文章不存在")
		return
	}

	// 获取当前用户ID和角色
	userID, exists := c.Get("user_id")
	if !exists {
		resp.FailWithCode(c, http.StatusUnauthorized, "未授权")
		return
	}

	// 检查权限 - 只有文章作者或管理员可以删除文章
	isAdmin, _ := c.Get("is_admin")
	isAuthor, err := ac.articleModel.CheckIsAuthor(uint(articleID), userID.(uint))
	if err != nil {
		logger.Error("检查文章作者失败", "article_id", articleID, "user_id", userID, "error", err)
		resp.FailWithMsg(c, "删除文章失败，请稍后重试")
		return
	}

	if !isAdmin.(bool) && !isAuthor {
		resp.FailWithCode(c, http.StatusForbidden, "只有文章作者或管理员可以删除文章")
		return
	}

	// 调用服务删除文章
	err = ac.articleModel.DeleteArticle(uint(articleID))
	if err != nil {
		logger.Error("删除文章失败", "article_id", articleID, "error", err)
		resp.FailWithMsg(c, "删除文章失败，请稍后重试")
		return
	}

	resp.OkWithMsg(c, "删除文章成功")
}

// ChangeArticleStatus 修改文章状态
// @Summary 修改文章状态
// @Description 修改文章状态（发布、下线等）
// @Tags 文章管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "文章ID"
// @Param status query int true "文章状态:1草稿,2待审核,3已发布,4已下线"
// @Success 200 {object} resp.Response "操作成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 404 {object} resp.Response "文章不存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/articles/{id}/status [put]
func (ac *ArticleController) ChangeArticleStatus(c *gin.Context) {
	// 获取路径参数
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil {
		resp.FailWithMsg(c, "无效的文章ID")
		return
	}

	// 获取状态参数
	status, err := strconv.Atoi(c.Query("status"))
	if err != nil || status < 1 || status > 4 {
		resp.FailWithMsg(c, "无效的文章状态")
		return
	}

	// 检查文章是否存在
	exists, err := ac.articleModel.CheckArticleExists(uint(articleID))
	if err != nil {
		logger.Error("检查文章是否存在失败", "article_id", articleID, "error", err)
		resp.FailWithMsg(c, "修改文章状态失败，请稍后重试")
		return
	}
	if !exists {
		resp.FailWithCode(c, http.StatusNotFound, "文章不存在")
		return
	}

	// 获取当前用户ID和角色
	userID, exists := c.Get("user_id")
	if !exists {
		resp.FailWithCode(c, http.StatusUnauthorized, "未授权")
		return
	}

	// 检查权限
	isAdmin, _ := c.Get("is_admin")
	isAuthor, err := ac.articleModel.CheckIsAuthor(uint(articleID), userID.(uint))
	if err != nil {
		logger.Error("检查文章作者失败", "article_id", articleID, "user_id", userID, "error", err)
		resp.FailWithMsg(c, "修改文章状态失败，请稍后重试")
		return
	}

	// 发布和下线操作需要管理员权限，作者只能将文章设为草稿或提交审核
	if (status == 3 || status == 4) && !isAdmin.(bool) {
		resp.FailWithCode(c, http.StatusForbidden, "只有管理员可以发布或下线文章")
		return
	}

	// 作者只能修改自己的文章
	if !isAdmin.(bool) && !isAuthor {
		resp.FailWithCode(c, http.StatusForbidden, "只能修改自己的文章")
		return
	}

	// 调用服务修改文章状态
	err = ac.articleModel.UpdateArticleStatus(uint(articleID), uint8(status))
	if err != nil {
		logger.Error("修改文章状态失败", "article_id", articleID, "status", status, "error", err)
		resp.FailWithMsg(c, "修改文章状态失败，请稍后重试")
		return
	}

	// 根据状态返回不同的消息
	var message string
	switch status {
	case 1:
		message = "文章已保存为草稿"
	case 2:
		message = "文章已提交审核"
	case 3:
		message = "文章已发布"
	case 4:
		message = "文章已下线"
	}

	resp.OkWithMsg(c, message)
}

// GetArticleContent 获取文章内容
// @Summary 获取文章内容
// @Description 获取文章内容
// @Tags 文章管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "文章ID"
// @Success 200 {object} resp.Response "返回文章内容"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 404 {object} resp.Response "文章不存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/articles/{id}/content [get]
func (ac *ArticleController) GetArticleContent(c *gin.Context) {
	// 获取路径参数
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil {
		resp.FailWithMsg(c, "无效的文章ID")
		return
	}

	// 检查文章是否存在
	exists, err := ac.articleModel.CheckArticleExists(uint(articleID))
	if err != nil {
		logger.Error("检查文章是否存在失败", "article_id", articleID, "error", err)
		resp.FailWithMsg(c, "获取文章内容失败，请稍后重试")
		return
	}
	if !exists {
		resp.FailWithCode(c, http.StatusNotFound, "文章不存在")
		return
	}

	// 获取文章内容
	content, err := ac.articleModel.GetArticleContent(uint(articleID))
	if err != nil {
		logger.Error("获取文章内容失败", "article_id", articleID, "error", err)
		resp.FailWithMsg(c, "获取文章内容失败")
		return
	}

	resp.OkWithData(c, gin.H{
		"content":        content.Content,
		"content_format": content.ContentFormat,
		"version":        content.Version,
		"updated_at":     content.UpdatedAt,
	})
}

// RegisterRoutes 注册路由
func (ac *ArticleController) RegisterRoutes(router *gin.RouterGroup) {
	articleGroup := router.Group("/articles")
	{
		// 文章列表
		articleGroup.GET("", middleware.RequirePermission("content:article:list"), ac.GetArticleList)

		// 文章详情
		articleGroup.GET("/:id", middleware.RequirePermission("content:article:info"), ac.GetArticleByID)

		// 文章内容
		articleGroup.GET("/:id/content", middleware.RequirePermission("content:article:info"), ac.GetArticleContent)

		// 创建文章
		articleGroup.POST("", middleware.RequirePermission("content:article:add"), ac.CreateArticle)

		// 更新文章
		articleGroup.PUT("/:id", middleware.RequirePermission("content:article:edit"), ac.UpdateArticle)

		// 删除文章
		articleGroup.DELETE("/:id", middleware.RequirePermission("content:article:delete"), ac.DeleteArticle)

		// 修改文章状态
		articleGroup.PUT("/:id/status", middleware.RequirePermission("content:article:edit"), ac.ChangeArticleStatus)
	}
}
