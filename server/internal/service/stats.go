package service

import (
	"time"

	"github.com/sunmoonstrand/go-react-blog/server/internal/model"
)

// SystemStats 系统统计数据
type SystemStats struct {
	UserCount     int `json:"user_count"`     // 用户数量
	ArticleCount  int `json:"article_count"`  // 文章数量
	CommentCount  int `json:"comment_count"`  // 评论数量
	CategoryCount int `json:"category_count"` // 分类数量
	TagCount      int `json:"tag_count"`      // 标签数量
	FileCount     int `json:"file_count"`     // 文件数量
	ViewCount     int `json:"view_count"`     // 总浏览量
	LikeCount     int `json:"like_count"`     // 总点赞量
}

// ArticleStats 文章统计数据
type ArticleStats struct {
	TotalArticles     int `json:"total_articles"`     // 文章总数
	PublishedArticles int `json:"published_articles"` // 已发布文章数
	DraftArticles     int `json:"draft_articles"`     // 草稿数
	TopArticles       int `json:"top_articles"`       // 置顶文章数
}

// UserStats 用户统计数据
type UserStats struct {
	TotalUsers      int `json:"total_users"`      // 用户总数
	ActiveUsers     int `json:"active_users"`     // 活跃用户数
	InactiveUsers   int `json:"inactive_users"`   // 未激活用户数
	AdminUsers      int `json:"admin_users"`      // 管理员数
	RegisteredToday int `json:"registered_today"` // 今日注册数
}

// GetSystemStats 获取系统统计数据
func GetSystemStats() (*SystemStats, error) {
	var stats SystemStats

	// 用户数量
	if err := model.DB.Model(&model.User{}).Count(&stats.UserCount).Error; err != nil {
		return nil, err
	}

	// 文章数量
	if err := model.DB.Model(&model.Article{}).Count(&stats.ArticleCount).Error; err != nil {
		return nil, err
	}

	// 评论数量
	if err := model.DB.Model(&model.Comment{}).Count(&stats.CommentCount).Error; err != nil {
		return nil, err
	}

	// 分类数量
	if err := model.DB.Model(&model.Category{}).Count(&stats.CategoryCount).Error; err != nil {
		return nil, err
	}

	// 标签数量
	if err := model.DB.Model(&model.Tag{}).Count(&stats.TagCount).Error; err != nil {
		return nil, err
	}

	// 文件数量
	if err := model.DB.Model(&model.File{}).Count(&stats.FileCount).Error; err != nil {
		return nil, err
	}

	// 总浏览量
	if err := model.DB.Model(&model.Article{}).Select("COALESCE(SUM(view_count), 0)").Scan(&stats.ViewCount).Error; err != nil {
		return nil, err
	}

	// 总点赞量
	if err := model.DB.Model(&model.Article{}).Select("COALESCE(SUM(like_count), 0)").Scan(&stats.LikeCount).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}

// GetArticleStats 获取文章统计数据
func GetArticleStats() (*ArticleStats, error) {
	var stats ArticleStats

	// 文章总数
	if err := model.DB.Model(&model.Article{}).Count(&stats.TotalArticles).Error; err != nil {
		return nil, err
	}

	// 已发布文章数
	if err := model.DB.Model(&model.Article{}).Where("is_published = ?", true).Count(&stats.PublishedArticles).Error; err != nil {
		return nil, err
	}

	// 草稿数
	if err := model.DB.Model(&model.Article{}).Where("is_published = ?", false).Count(&stats.DraftArticles).Error; err != nil {
		return nil, err
	}

	// 置顶文章数
	if err := model.DB.Model(&model.Article{}).Where("is_top = ?", true).Count(&stats.TopArticles).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}

// GetUserStats 获取用户统计数据
func GetUserStats() (*UserStats, error) {
	var stats UserStats

	// 用户总数
	if err := model.DB.Model(&model.User{}).Count(&stats.TotalUsers).Error; err != nil {
		return nil, err
	}

	// 活跃用户数
	if err := model.DB.Model(&model.User{}).Where("is_enabled = ?", true).Count(&stats.ActiveUsers).Error; err != nil {
		return nil, err
	}

	// 未激活用户数
	if err := model.DB.Model(&model.User{}).Where("is_enabled = ?", false).Count(&stats.InactiveUsers).Error; err != nil {
		return nil, err
	}

	// 管理员数
	if err := model.DB.Table("sys_users").
		Joins("JOIN sys_user_roles ON sys_users.user_id = sys_user_roles.user_id").
		Joins("JOIN sys_roles ON sys_user_roles.role_id = sys_roles.role_id").
		Where("sys_roles.role_key = ?", "admin").
		Count(&stats.AdminUsers).Error; err != nil {
		return nil, err
	}

	// 今日注册数
	today := time.Now().Format("2006-01-02")
	if err := model.DB.Model(&model.User{}).
		Where("DATE(created_at) = ?", today).
		Count(&stats.RegisteredToday).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}

// GetRecentArticles 获取最近文章
func GetRecentArticles(limit int) ([]model.ArticleResponse, error) {
	var articles []model.Article
	if err := model.DB.Where("is_published = ?", true).
		Preload("Category").
		Preload("User", func(db *model.DB) *model.DB {
			return db.Select("user_id, username, nickname, avatar")
		}).
		Preload("Tags").
		Order("published_at DESC").
		Limit(limit).
		Find(&articles).Error; err != nil {
		return nil, err
	}

	// 转换为响应对象
	var articleResponses []model.ArticleResponse
	for _, article := range articles {
		response := model.ArticleResponse{
			ArticleID:    article.ArticleID,
			Title:        article.Title,
			Summary:      article.Summary,
			CategoryID:   article.CategoryID,
			CategoryName: article.Category.CategoryName,
			CoverImage:   article.CoverImage,
			IsTop:        article.IsTop,
			IsPublished:  article.IsPublished,
			ViewCount:    article.ViewCount,
			LikeCount:    article.LikeCount,
			CommentCount: article.CommentCount,
			UserID:       article.UserID,
			Username:     article.User.Username,
			Nickname:     article.User.Nickname,
			Avatar:       article.User.Avatar,
			PublishedAt:  article.PublishedAt,
			CreatedAt:    article.CreatedAt,
			UpdatedAt:    article.UpdatedAt,
		}

		// 添加标签
		for _, tag := range article.Tags {
			response.Tags = append(response.Tags, model.TagInfo{
				TagID:   tag.TagID,
				TagName: tag.TagName,
				TagKey:  tag.TagKey,
			})
		}

		articleResponses = append(articleResponses, response)
	}

	return articleResponses, nil
}

// GetRecentComments 获取最近评论
func GetRecentComments(limit int) ([]model.CommentResponse, error) {
	var comments []model.Comment
	if err := model.DB.Where("is_visible = ?", true).
		Preload("User", func(db *model.DB) *model.DB {
			return db.Select("user_id, username, nickname, avatar")
		}).
		Preload("Article", func(db *model.DB) *model.DB {
			return db.Select("article_id, title")
		}).
		Order("created_at DESC").
		Limit(limit).
		Find(&comments).Error; err != nil {
		return nil, err
	}

	// 转换为响应对象
	var commentResponses []model.CommentResponse
	for _, comment := range comments {
		commentResponses = append(commentResponses, model.CommentResponse{
			CommentID:    comment.CommentID,
			ArticleID:    comment.ArticleID,
			UserID:       comment.UserID,
			Username:     comment.User.Username,
			Nickname:     comment.User.Nickname,
			Avatar:       comment.User.Avatar,
			Content:      comment.Content,
			ParentID:     comment.ParentID,
			IsVisible:    comment.IsVisible,
			CreatedAt:    comment.CreatedAt,
			UpdatedAt:    comment.UpdatedAt,
			ArticleTitle: comment.Article.Title,
		})
	}

	return commentResponses, nil
}

// GetPopularArticles 获取热门文章
func GetPopularArticles(limit int) ([]model.ArticleResponse, error) {
	var articles []model.Article
	if err := model.DB.Where("is_published = ?", true).
		Preload("Category").
		Preload("User", func(db *model.DB) *model.DB {
			return db.Select("user_id, username, nickname, avatar")
		}).
		Preload("Tags").
		Order("view_count DESC").
		Limit(limit).
		Find(&articles).Error; err != nil {
		return nil, err
	}

	// 转换为响应对象
	var articleResponses []model.ArticleResponse
	for _, article := range articles {
		response := model.ArticleResponse{
			ArticleID:    article.ArticleID,
			Title:        article.Title,
			Summary:      article.Summary,
			CategoryID:   article.CategoryID,
			CategoryName: article.Category.CategoryName,
			CoverImage:   article.CoverImage,
			IsTop:        article.IsTop,
			IsPublished:  article.IsPublished,
			ViewCount:    article.ViewCount,
			LikeCount:    article.LikeCount,
			CommentCount: article.CommentCount,
			UserID:       article.UserID,
			Username:     article.User.Username,
			Nickname:     article.User.Nickname,
			Avatar:       article.User.Avatar,
			PublishedAt:  article.PublishedAt,
			CreatedAt:    article.CreatedAt,
			UpdatedAt:    article.UpdatedAt,
		}

		// 添加标签
		for _, tag := range article.Tags {
			response.Tags = append(response.Tags, model.TagInfo{
				TagID:   tag.TagID,
				TagName: tag.TagName,
				TagKey:  tag.TagKey,
			})
		}

		articleResponses = append(articleResponses, response)
	}

	return articleResponses, nil
}
