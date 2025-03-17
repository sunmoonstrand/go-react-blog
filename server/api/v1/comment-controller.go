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

// CommentController 评论控制器
type CommentController struct {
	commentModel *model.CommentModel
	articleModel *model.ArticleModel
}

// NewCommentController 创建评论控制器实例
func NewCommentController(commentModel *model.CommentModel, articleModel *model.ArticleModel) *CommentController {
	return &CommentController{
		commentModel: commentModel,
		articleModel: articleModel,
	}
}

// CreateCommentRequest 创建评论请求结构
type CreateCommentRequest struct {
	ArticleID uint   `json:"article_id" binding:"required"`
	ParentID  uint   `json:"parent_id"`
	Content   string `json:"content" binding:"required,min=2,max=500"`
}

// UpdateCommentRequest 更新评论请求结构
type UpdateCommentRequest struct {
	Content    string `json:"content" binding:"required,min=2,max=500"`
	IsApproved *bool  `json:"is_approved"`
}

// ApproveCommentRequest 审核评论请求结构
type ApproveCommentRequest struct {
	IsApproved bool `json:"is_approved" binding:"required"`
}

// GetCommentList 获取评论列表
// @Summary 获取评论列表
// @Description 分页获取评论列表
// @Tags 评论管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码，默认1" default(1)
// @Param page_size query int false "每页记录数，默认20" default(20)
// @Param article_id query int false "文章ID筛选"
// @Param user_id query int false "用户ID筛选"
// @Param approved query int false "审核状态:0全部,1待审核,2已通过,3已拒绝"
// @Success 200 {object} resp.Response{data=resp.PageResult} "返回评论列表"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/comments [get]
func (cc *CommentController) GetCommentList(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 获取筛选参数
	articleID, _ := strconv.Atoi(c.DefaultQuery("article_id", "0"))
	userID, _ := strconv.Atoi(c.DefaultQuery("user_id", "0"))
	approved, _ := strconv.Atoi(c.DefaultQuery("approved", "0"))

	// 调用服务获取评论列表
	comments, total, err := cc.commentModel.GetCommentList(page, pageSize, uint(articleID), uint(userID), uint8(approved))
	if err != nil {
		logger.Error("获取评论列表失败", "error", err)
		resp.FailWithMsg(c, "获取评论列表失败")
		return
	}

	resp.OkWithPage(c, comments, total, page, pageSize)
}

// GetArticleComments 获取文章评论
// @Summary 获取文章评论
// @Description 获取指定文章的评论列表
// @Tags 评论管理
// @Accept json
// @Produce json
// @Param article_id path int true "文章ID"
// @Param page query int false "页码，默认1" default(1)
// @Param page_size query int false "每页记录数，默认20" default(20)
// @Success 200 {object} resp.Response{data=resp.PageResult} "返回评论列表"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 404 {object} resp.Response "文章不存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/articles/{article_id}/comments [get]
func (cc *CommentController) GetArticleComments(c *gin.Context) {
	// 获取路径参数
	articleID, err := strconv.ParseUint(c.Param("article_id"), 10, 0)
	if err != nil {
		resp.FailWithMsg(c, "无效的文章ID")
		return
	}

	// 检查文章是否存在
	exists, err := cc.articleModel.CheckArticleExists(uint(articleID))
	if err != nil {
		logger.Error("检查文章是否存在失败", "article_id", articleID, "error", err)
		resp.FailWithMsg(c, "获取评论失败，请稍后重试")
		return
	}
	if !exists {
		resp.FailWithCode(c, http.StatusNotFound, "文章不存在")
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 调用服务获取文章评论
	comments, total, err := cc.commentModel.GetArticleComments(uint(articleID), page, pageSize)
	if err != nil {
		logger.Error("获取文章评论失败", "article_id", articleID, "error", err)
		resp.FailWithMsg(c, "获取评论失败")
		return
	}

	resp.OkWithPage(c, comments, total, page, pageSize)
}

// GetCommentByID 根据ID获取评论详情
// @Summary 获取评论详情
// @Description 根据评论ID获取评论详细信息
// @Tags 评论管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "评论ID"
// @Success 200 {object} resp.Response "返回评论详情"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 404 {object} resp.Response "评论不存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/comments/{id} [get]
func (cc *CommentController) GetCommentByID(c *gin.Context) {
	// 获取路径参数
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil {
		resp.FailWithMsg(c, "无效的评论ID")
		return
	}

	// 获取评论详情
	comment, err := cc.commentModel.GetCommentByID(uint(commentID))
	if err != nil {
		if err.Error() == "record not found" {
			resp.FailWithCode(c, http.StatusNotFound, "评论不存在")
		} else {
			logger.Error("获取评论详情失败", "comment_id", commentID, "error", err)
			resp.FailWithMsg(c, "获取评论详情失败")
		}
		return
	}

	// 获取评论回复
	replies, err := cc.commentModel.GetCommentReplies(uint(commentID))
	if err != nil {
		logger.Error("获取评论回复失败", "comment_id", commentID, "error", err)
		// 这里不返回错误，只是回复可能为空
	}

	// 构建响应数据
	respData := gin.H{
		"comment_id":      comment.ID,
		"article_id":      comment.ArticleID,
		"user_id":         comment.UserID,
		"user_info":       comment.UserInfo,
		"parent_id":       comment.ParentID,
		"root_id":         comment.RootID,
		"content":         comment.Content,
		"liked_count":     comment.LikedCount,
		"is_approved":     comment.IsApproved,
		"is_admin_reply":  comment.IsAdminReply,
		"created_at":      comment.CreatedAt,
		"updated_at":      comment.UpdatedAt,
		"replies":         replies,
	}

	resp.OkWithData(c, respData)
}

// CreateComment 创建评论
// @Summary 创建评论
// @Description 创建新评论
// @Tags 评论管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body CreateCommentRequest true "评论信息"
// @Success 200 {object} resp.Response "创建成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 404 {object} resp.Response "文章不存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/comments [post]
func (cc *CommentController) CreateComment(c *gin.Context) {
	var req CreateCommentRequest
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

	// 检查文章是否存在
	exists, err := cc.articleModel.CheckArticleExists(req.ArticleID)
	if err != nil {
		logger.Error("检查文章是否存在失败", "article_id", req.ArticleID, "error", err)
		resp.FailWithMsg(c, "创建评论失败，请稍后重试")
		return
	}
	if !exists {
		resp.FailWithCode(c, http.StatusNotFound, "文章不存在")
		return
	}

	// 检查文章是否允许评论
	allowComment, err := cc.articleModel.CheckAllowComment(req.ArticleID)
	if err != nil {
		logger.Error("检查文章是否允许评论失败", "article_id", req.ArticleID, "error", err)
		resp.FailWithMsg(c, "创建评论失败，请稍后重试")
		return
	}
	if !allowComment {
		resp.FailWithMsg(c, "该文章不允许评论")
		return
	}

	// 如果有父评论，检查父评论是否存在
	var rootID uint
	if req.ParentID > 0 {
		parentComment, err := cc.commentModel.GetCommentByID(req.ParentID)
		if err != nil {
			if err.Error() == "record not found" {
				resp.FailWithCode(c, http.StatusNotFound, "父评论不存在")
			} else {
				logger.Error("获取父评论失败", "parent_id", req.ParentID, "error", err)
				resp.FailWithMsg(c, "创建评论失败，请稍后重试")
			}
			return
		}

		// 检查父评论是否属于同一篇文章
		if parentComment.ArticleID != req.ArticleID {
			resp.FailWithMsg(c, "父评论与当前文章不匹配")
			return
		}

		// 设置根评论ID
		if parentComment.RootID > 0 {
			rootID = parentComment.RootID
		} else {
			rootID = parentComment.ID
		}
	}

	// 判断用户角色，确定是否为管理员回复
	isAdmin, _ := c.Get("is_admin")
	isAdminReply := isAdmin.(bool)

	// 创建评论
	comment := &model.Comment{
		ArticleID:    req.ArticleID,
		UserID:       userID.(uint),
		ParentID:     req.ParentID,
		RootID:       rootID,
		Content:      req.Content,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.GetHeader("User-Agent"),
		IsApproved:   isAdminReply, // 管理员评论自动通过审核
		IsAdminReply: isAdminReply,
	}

	if err := cc.commentModel.CreateComment(comment); err != nil {
		logger.Error("创建评论失败", "error", err)
		resp.FailWithMsg(c, "创建评论失败，请稍后重试")
		return
	}

	resp.OkWithData(c, gin.H{
		"comment_id": comment.ID,
		"message":    isAdminReply ? "评论发表成功" : "评论发表成功，等待审核",
	})
}

// UpdateComment 更新评论
// @Summary 更新评论
// @Description 更新评论内容
// @Tags 评论管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "评论ID"
// @Param data body UpdateCommentRequest true "评论信息"
// @Success 200 {object} resp.Response "更新成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 404 {object} resp.Response "评论不存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/comments/{id} [put]
func (cc *CommentController) UpdateComment(c *gin.Context) {
	// 获取路径参数
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil {
		resp.FailWithMsg(c, "无效的评论ID")
		return
	}

	var req UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.FailWithValidation(c, err)
		return
	}

	// 获取评论信息
	comment, err := cc.commentModel.GetCommentByID(uint(commentID))
	if err != nil {
		if err.Error() == "record not found" {
			resp.FailWithCode(c, http.StatusNotFound, "评论不存在")
		} else {
			logger.Error("获取评论信息失败", "comment_id", commentID, "error", err)
			resp.FailWithMsg(c, "更新评论失败，请稍后重试")
		}
		return
	}

	// 获取当前用户ID和角色
	userID, exists := c.Get("user_id")
	if !exists {
		resp.FailWithCode(c, http.StatusUnauthorized, "未授权")
		return
	}

	// 检查权限 - 只有评论作者或管理员可以修改评论
	isAdmin, _ := c.Get("is_admin")
	if !isAdmin.(bool) && comment.UserID != userID.(uint) {
		resp.FailWithCode(c, http.StatusForbidden, "只有评论作者或管理员可以修改评论")
		return
	}

	// 构建评论更新信息
	updates := make(map[string]interface{})
	updates["content"] = req.Content

	// 管理员可以更新审核状态
	if isAdmin.(bool) && req.IsApproved != nil {
		updates["is_approved"] = *req.IsApproved
	}

	// 调用服务更新评论
	if err := cc.commentModel.UpdateComment(uint(commentID), updates); err != nil {
		logger.Error("更新评论失败", "comment_id", commentID, "error", err)
		resp.FailWithMsg(c, "更新评论失败，请稍后重试")
		return
	}

	resp.OkWithMsg(c, "更新评论成功")
}

// DeleteComment 删除评论
// @Summary 删除评论
// @Description 根据ID删除评论
// @Tags 评论管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "评论ID"
// @Success 200 {object} resp.Response "删除成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 404 {object} resp.Response "评论不存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/comments/{id} [delete]
func (cc *CommentController) DeleteComment(c *gin.Context) {
	// 获取路径参数
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil {
		resp.FailWithMsg(c, "无效的评论ID")
		return
	}

	// 获取评论信息
	comment, err := cc.commentModel.GetCommentByID(uint(commentID))
	if err != nil {
		if err.Error() == "record not found" {
			resp.FailWithCode(c, http.StatusNotFound, "评论不存在")
		} else {
			logger.Error("获取评论信息失败", "comment_id", commentID, "error", err)
			resp.FailWithMsg(c, "删除评论失败，请稍后重试")
		}
		return
	}

	// 获取当前用户ID和角色
	userID, exists := c.Get("user_id")
	if !exists {
		resp.FailWithCode(c, http.StatusUnauthorized, "未授权")
		return
	}

	// 检查权限 - 只有评论作者或管理员可以删除评论
	isAdmin, _ := c.Get("is_admin")
	if !isAdmin.(bool) && comment.UserID != userID.(uint) {
		resp.FailWithCode(c, http.StatusForbidden, "只有评论作者或管理员可以删除评论")
		return
	}

	// 调用服务删除评论
	if err := cc.commentModel.DeleteComment(uint(commentID)); err != nil {
		logger.Error("删除评论失败", "comment_id", commentID, "error", err)
		resp.FailWithMsg(c, "删除评论失败，请稍后重试")
		return
	}

	resp.OkWithMsg(c, "删除评论成功")
}

// ApproveComment 审核评论
// @Summary 审核评论
// @Description 审核通过或拒绝评论
// @Tags 评论管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "评论ID"
// @Param data body ApproveCommentRequest true "审核信息"
// @Success 200 {object} resp.Response "审核成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 404 {object} resp.Response "评论不存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/comments/{id}/approve [put]
func (cc *CommentController) ApproveComment(c *gin.Context) {
	// 获取路径参数
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil {
		resp.FailWithMsg(c, "无效的评论ID")
		return
	}

	var req ApproveCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.FailWithValidation(c, err)
		return
	}

	// 检查评论是否存在
	exists, err := cc.commentModel.CheckCommentExists(uint(commentID))
	if err != nil {
		logger.Error("检查评论是否存在失败", "comment_id", commentID, "error", err)
		resp.FailWithMsg(c, "审核评论失败，请稍后重试")
		return
	}
	if !exists {
		resp.FailWithCode(c, http.StatusNotFound, "评论不存在")
		return
	}

	// 调用服务审核评论
	if err := cc.commentModel.ApproveComment(uint(commentID), req.IsApproved); err != nil {
		logger.Error("审核评论失败", "comment_id", commentID, "approved", req.IsApproved, "error", err)
		resp.FailWithMsg(c, "审核评论失败，请稍后重试")
		return
	}

	resp.OkWithMsg(c, req.IsApproved ? "评论已通过审核" : "评论已被拒绝")
}

// BatchApproveComments 批量审核评论
// @Summary 批量审核评论
// @Description 批量审核通过或拒绝多个评论
// @Tags 评论管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param is_approved query bool true "是否通过审核"
// @Param ids body []int true "评论ID数组"
// @Success 200 {object} resp.Response "审核成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/comments/batch/approve [put]
func (cc *CommentController) BatchApproveComments(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		resp.FailWithMsg(c, "无效的参数")
		return
	}

	if len(ids) == 0 {
		resp.FailWithMsg(c, "请选择要审核的评论")
		return
	}

	// 获取审核状态
	isApproved, err := strconv.ParseBool(c.Query("is_approved"))
	if err != nil {
		resp.FailWithMsg(c, "无效的审核状态")
		return
	}

	// 调用服务批量审核评论
	result, err := cc.commentModel.BatchApproveComments(ids, isApproved)
	if err != nil {
		logger.Error("批量审核评论失败", "error", err)
		resp.FailWithMsg(c, "批量审核评论失败")
		return
	}

	var message string
	if isApproved {
		message = "批量通过评论完成"
	} else {
		message = "批量拒绝评论完成"
	}

	resp.OkWithData(c, gin.H{
		"success": result.Success,
		"failed":  result.Failed,
		"message": message,
	})
}

// RegisterRoutes 注册路由
func (cc *CommentController) RegisterRoutes(router *gin.RouterGroup) {
	commentGroup := router.Group("/comments")
	{
		// 获取评论列表(后台管理)
		commentGroup.GET("", middleware.RequirePermission("content:comment:list"), cc.GetCommentList)

		// 创建评论
		commentGroup.POST("", middleware.RequireAuth(), cc.CreateComment)

		// 获取评论详情
		commentGroup.GET("/:id", middleware.RequirePermission("content:comment:info"), cc.GetCommentByID)

		// 更新评论
		commentGroup.PUT("/:id", middleware.RequireAuth(), cc.UpdateComment)

		// 删除评论
		commentGroup.DELETE("/:id", middleware.RequireAuth(), cc.DeleteComment)

		// 审核评论
		commentGroup.PUT("/:id/approve", middleware.RequirePermission("content:comment:edit"), cc.ApproveComment)

		// 批量审核评论
		commentGroup.PUT("/batch/approve", middleware.RequirePermission("content:comment:edit"), cc.BatchApproveComments)
	}

	// 获取文章评论(前台展示)
	articleGroup := router.Group("/articles")
	articleGroup.GET("/:article_id/comments", cc.GetArticleComments)
}