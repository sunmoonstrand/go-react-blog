package service

import (
	"errors"

	"github.com/sunmoonstrand/go-react-blog/server/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CreateComment 创建评论
func CreateComment(form model.CommentCreateForm, userID int) (int, error) {
	// 检查文章是否存在
	var article model.Article
	if err := model.DB.First(&article, form.ArticleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("文章不存在")
		}
		return 0, err
	}

	// 检查文章是否已发布
	if !article.IsPublished {
		return 0, errors.New("文章未发布，不能评论")
	}

	// 创建评论
	comment := model.Comment{
		ArticleID: form.ArticleID,
		UserID:    userID,
		Content:   form.Content,
		ParentID:  form.ParentID,
		IsVisible: true, // 默认可见
	}

	// 开启事务
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 保存评论
	if err := tx.Create(&comment).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	// 更新文章评论数
	if err := tx.Model(&article).Update("comment_count", gorm.Expr("comment_count + ?", 1)).Error; err != nil {
		tx.Rollback()
		zap.L().Error("更新文章评论数失败",
			zap.Int("article_id", form.ArticleID),
			zap.Error(err),
		)
		// 不返回错误，因为评论已创建成功
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return 0, err
	}

	return comment.CommentID, nil
}

// UpdateComment 更新评论
func UpdateComment(commentID int, content string, userID int) error {
	// 检查评论是否存在
	var comment model.Comment
	if err := model.DB.First(&comment, commentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("评论不存在")
		}
		return err
	}

	// 检查是否有权限更新（只有评论作者或管理员可以更新）
	if comment.UserID != userID {
		// 检查是否为管理员
		var isAdmin bool
		var count int64
		if err := model.DB.Table("sys_user_roles").
			Joins("JOIN sys_roles ON sys_user_roles.role_id = sys_roles.role_id").
			Where("sys_user_roles.user_id = ? AND sys_roles.role_key = ?", userID, "admin").
			Count(&count).Error; err != nil {
			return err
		}
		isAdmin = count > 0
		if !isAdmin {
			return errors.New("无权限更新该评论")
		}
	}

	// 更新评论
	if err := model.DB.Model(&comment).Update("content", content).Error; err != nil {
		return err
	}

	return nil
}

// DeleteComment 删除评论
func DeleteComment(commentID int, userID int) error {
	// 检查评论是否存在
	var comment model.Comment
	if err := model.DB.First(&comment, commentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("评论不存在")
		}
		return err
	}

	// 检查是否有权限删除（只有评论作者或管理员可以删除）
	if comment.UserID != userID {
		// 检查是否为管理员
		var isAdmin bool
		var count int64
		if err := model.DB.Table("sys_user_roles").
			Joins("JOIN sys_roles ON sys_user_roles.role_id = sys_roles.role_id").
			Where("sys_user_roles.user_id = ? AND sys_roles.role_key = ?", userID, "admin").
			Count(&count).Error; err != nil {
			return err
		}
		isAdmin = count > 0
		if !isAdmin {
			return errors.New("无权限删除该评论")
		}
	}

	// 开启事务
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除评论
	if err := tx.Delete(&comment).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 更新文章评论数
	if err := tx.Model(&model.Article{}).Where("article_id = ?", comment.ArticleID).
		Update("comment_count", gorm.Expr("comment_count - ?", 1)).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// GetCommentByID 根据ID获取评论
func GetCommentByID(commentID int) (*model.Comment, error) {
	var comment model.Comment
	if err := model.DB.First(&comment, commentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("评论不存在")
		}
		return nil, err
	}
	return &comment, nil
}

// ListComments 获取评论列表
func ListComments(params model.CommentQueryParams) (*model.PageResult, error) {
	var comments []model.Comment
	var total int64

	// 构建查询
	query := model.DB.Model(&model.Comment{}).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("user_id, username, nickname, avatar")
		})

	// 应用过滤条件
	if params.ArticleID != nil {
		query = query.Where("article_id = ?", *params.ArticleID)
	}
	if params.UserID != nil {
		query = query.Where("user_id = ?", *params.UserID)
	}
	if params.ParentID != nil {
		query = query.Where("parent_id = ?", *params.ParentID)
	}
	if params.IsVisible != nil {
		query = query.Where("is_visible = ?", *params.IsVisible)
	}
	if params.Keyword != "" {
		query = query.Where("content LIKE ?", "%"+params.Keyword+"%")
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
	if err := query.Offset(offset).Limit(params.PageSize).Order("comment_id DESC").Find(&comments).Error; err != nil {
		return nil, err
	}

	// 转换为响应对象
	var commentResponses []model.CommentResponse
	for _, comment := range comments {
		commentResponses = append(commentResponses, model.CommentResponse{
			CommentID: comment.CommentID,
			ArticleID: comment.ArticleID,
			UserID:    comment.UserID,
			Username:  comment.User.Username,
			Nickname:  comment.User.Nickname,
			Avatar:    comment.User.Avatar,
			Content:   comment.Content,
			ParentID:  comment.ParentID,
			IsVisible: comment.IsVisible,
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
		})
	}

	return model.NewPageResult(commentResponses, total, params.Page, params.PageSize), nil
}

// UpdateCommentVisibility 更新评论可见性
func UpdateCommentVisibility(commentID int, isVisible bool) error {
	// 检查评论是否存在
	var comment model.Comment
	if err := model.DB.First(&comment, commentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("评论不存在")
		}
		return err
	}

	// 更新可见性
	if err := model.DB.Model(&comment).Update("is_visible", isVisible).Error; err != nil {
		return err
	}

	return nil
}

// GetArticleComments 获取文章评论（树形结构）
func GetArticleComments(articleID int) ([]model.CommentResponse, error) {
	var comments []model.Comment
	if err := model.DB.Where("article_id = ? AND is_visible = ?", articleID, true).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("user_id, username, nickname, avatar")
		}).
		Order("created_at ASC").
		Find(&comments).Error; err != nil {
		return nil, err
	}

	// 转换为响应对象
	var commentResponses []model.CommentResponse
	for _, comment := range comments {
		commentResponses = append(commentResponses, model.CommentResponse{
			CommentID: comment.CommentID,
			ArticleID: comment.ArticleID,
			UserID:    comment.UserID,
			Username:  comment.User.Username,
			Nickname:  comment.User.Nickname,
			Avatar:    comment.User.Avatar,
			Content:   comment.Content,
			ParentID:  comment.ParentID,
			IsVisible: comment.IsVisible,
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
		})
	}

	return commentResponses, nil
}
