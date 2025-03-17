package service

import (
	"errors"
	"time"

	"github.com/sunmoonstrand/go-react-blog/server/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CreateArticle 创建文章
func CreateArticle(form model.ArticleCreateForm, userID int) (int, error) {
	// 检查分类是否存在
	var category model.Category
	if err := model.DB.First(&category, form.CategoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("分类不存在")
		}
		return 0, err
	}

	// 创建文章
	article := model.Article{
		Title:        form.Title,
		Content:      form.Content,
		Summary:      form.Summary,
		CategoryID:   form.CategoryID,
		CoverImage:   form.CoverImage,
		IsTop:        form.IsTop,
		IsPublished:  form.IsPublished,
		ViewCount:    0,
		LikeCount:    0,
		CommentCount: 0,
		UserID:       userID,
		PublishedAt:  nil,
	}

	// 如果发布，设置发布时间
	if form.IsPublished {
		now := time.Now()
		article.PublishedAt = &now
	}

	// 开启事务
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 保存文章
	if err := tx.Create(&article).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	// 保存标签关联
	if len(form.TagIDs) > 0 {
		for _, tagID := range form.TagIDs {
			// 检查标签是否存在
			var tag model.Tag
			if err := tx.First(&tag, tagID).Error; err != nil {
				tx.Rollback()
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return 0, errors.New("标签不存在")
				}
				return 0, err
			}

			// 创建关联
			articleTag := model.ArticleTag{
				ArticleID: article.ArticleID,
				TagID:     tagID,
			}
			if err := tx.Create(&articleTag).Error; err != nil {
				tx.Rollback()
				return 0, err
			}
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return 0, err
	}

	return article.ArticleID, nil
}

// UpdateArticle 更新文章
func UpdateArticle(articleID int, form model.ArticleUpdateForm, userID int) error {
	// 检查文章是否存在
	var article model.Article
	if err := model.DB.First(&article, articleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("文章不存在")
		}
		return err
	}

	// 检查是否有权限更新（只有作者或管理员可以更新）
	if article.UserID != userID {
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
			return errors.New("无权限更新该文章")
		}
	}

	// 检查分类是否存在
	if form.CategoryID != nil {
		var category model.Category
		if err := model.DB.First(&category, *form.CategoryID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("分类不存在")
			}
			return err
		}
	}

	// 开启事务
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新文章
	updates := map[string]interface{}{}
	if form.Title != "" {
		updates["title"] = form.Title
	}
	if form.Content != "" {
		updates["content"] = form.Content
	}
	if form.Summary != nil {
		updates["summary"] = *form.Summary
	}
	if form.CategoryID != nil {
		updates["category_id"] = *form.CategoryID
	}
	if form.CoverImage != nil {
		updates["cover_image"] = *form.CoverImage
	}
	if form.IsTop != nil {
		updates["is_top"] = *form.IsTop
	}
	if form.IsPublished != nil {
		updates["is_published"] = *form.IsPublished
		// 如果从未发布变为发布，设置发布时间
		if *form.IsPublished && article.PublishedAt == nil {
			now := time.Now()
			updates["published_at"] = now
		}
	}

	if err := tx.Model(&article).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 更新标签关联
	if form.TagIDs != nil {
		// 删除现有标签关联
		if err := tx.Where("article_id = ?", articleID).Delete(&model.ArticleTag{}).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 添加新标签关联
		for _, tagID := range *form.TagIDs {
			// 检查标签是否存在
			var tag model.Tag
			if err := tx.First(&tag, tagID).Error; err != nil {
				tx.Rollback()
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("标签不存在")
				}
				return err
			}

			// 创建关联
			articleTag := model.ArticleTag{
				ArticleID: articleID,
				TagID:     tagID,
			}
			if err := tx.Create(&articleTag).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// GetArticleByID 根据ID获取文章
func GetArticleByID(articleID int, increaseViewCount bool) (*model.ArticleResponse, error) {
	var article model.Article
	if err := model.DB.Preload("Category").
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("user_id, username, nickname, avatar")
		}).
		Preload("Tags").
		First(&article, articleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("文章不存在")
		}
		return nil, err
	}

	// 增加浏览量
	if increaseViewCount {
		if err := model.DB.Model(&article).Update("view_count", gorm.Expr("view_count + ?", 1)).Error; err != nil {
			zap.L().Error("增加文章浏览量失败",
				zap.Int("article_id", articleID),
				zap.Error(err),
			)
			// 不返回错误，因为获取文章已成功
		}
		article.ViewCount++
	}

	// 转换为响应对象
	response := model.ArticleResponse{
		ArticleID:    article.ArticleID,
		Title:        article.Title,
		Content:      article.Content,
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

	return &response, nil
}

// ListArticles 获取文章列表
func ListArticles(params model.ArticleQueryParams) (*model.PageResult, error) {
	var articles []model.Article
	var total int64

	// 构建查询
	query := model.DB.Model(&model.Article{}).
		Preload("Category").
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("user_id, username, nickname, avatar")
		}).
		Preload("Tags")

	// 应用过滤条件
	if params.Title != "" {
		query = query.Where("title LIKE ?", "%"+params.Title+"%")
	}
	if params.CategoryID != nil {
		query = query.Where("category_id = ?", *params.CategoryID)
	}
	if params.UserID != nil {
		query = query.Where("user_id = ?", *params.UserID)
	}
	if params.IsPublished != nil {
		query = query.Where("is_published = ?", *params.IsPublished)
	}
	if params.IsTop != nil {
		query = query.Where("is_top = ?", *params.IsTop)
	}
	if params.TagID != nil {
		query = query.Joins("JOIN cms_article_tags ON cms_articles.article_id = cms_article_tags.article_id").
			Where("cms_article_tags.tag_id = ?", *params.TagID)
	}
	if params.StartTime != nil && params.EndTime != nil {
		query = query.Where("created_at BETWEEN ? AND ?", params.StartTime, params.EndTime)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 排序
	orderBy := "created_at DESC"
	if params.OrderBy != "" {
		switch params.OrderBy {
		case "view_count":
			orderBy = "view_count DESC"
		case "like_count":
			orderBy = "like_count DESC"
		case "comment_count":
			orderBy = "comment_count DESC"
		case "published_at":
			orderBy = "published_at DESC"
		}
	}

	// 置顶文章优先
	query = query.Order("is_top DESC").Order(orderBy)

	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	if err := query.Offset(offset).Limit(params.PageSize).Find(&articles).Error; err != nil {
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

	return model.NewPageResult(articleResponses, total, params.Page, params.PageSize), nil
}

// DeleteArticle 删除文章
func DeleteArticle(articleID int, userID int) error {
	// 检查文章是否存在
	var article model.Article
	if err := model.DB.First(&article, articleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("文章不存在")
		}
		return err
	}

	// 检查是否有权限删除（只有作者或管理员可以删除）
	if article.UserID != userID {
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
			return errors.New("无权限删除该文章")
		}
	}

	// 开启事务
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除文章标签关联
	if err := tx.Where("article_id = ?", articleID).Delete(&model.ArticleTag{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 删除文章评论
	if err := tx.Where("article_id = ?", articleID).Delete(&model.Comment{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 删除文章
	if err := tx.Delete(&article).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// UpdateArticleStatus 更新文章状态
func UpdateArticleStatus(articleID int, isPublished bool, userID int) error {
	// 检查文章是否存在
	var article model.Article
	if err := model.DB.First(&article, articleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("文章不存在")
		}
		return err
	}

	// 检查是否有权限更新（只有作者或管理员可以更新）
	if article.UserID != userID {
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
			return errors.New("无权限更新该文章")
		}
	}

	// 更新状态
	updates := map[string]interface{}{
		"is_published": isPublished,
	}

	// 如果从未发布变为发布，设置发布时间
	if isPublished && article.PublishedAt == nil {
		now := time.Now()
		updates["published_at"] = now
	}

	if err := model.DB.Model(&article).Updates(updates).Error; err != nil {
		return err
	}

	return nil
}

// UpdateArticleTop 更新文章置顶状态
func UpdateArticleTop(articleID int, isTop bool, userID int) error {
	// 检查文章是否存在
	var article model.Article
	if err := model.DB.First(&article, articleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("文章不存在")
		}
		return err
	}

	// 检查是否有权限更新（只有作者或管理员可以更新）
	if article.UserID != userID {
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
			return errors.New("无权限更新该文章")
		}
	}

	// 更新置顶状态
	if err := model.DB.Model(&article).Update("is_top", isTop).Error; err != nil {
		return err
	}

	return nil
}

// LikeArticle 点赞文章
func LikeArticle(articleID int, userID int) error {
	// 检查文章是否存在
	var article model.Article
	if err := model.DB.First(&article, articleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("文章不存在")
		}
		return err
	}

	// 检查是否已点赞
	var count int64
	if err := model.DB.Model(&model.ArticleLike{}).
		Where("article_id = ? AND user_id = ?", articleID, userID).
		Count(&count).Error; err != nil {
		return err
	}

	// 已点赞则取消点赞，未点赞则添加点赞
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if count > 0 {
		// 取消点赞
		if err := tx.Where("article_id = ? AND user_id = ?", articleID, userID).Delete(&model.ArticleLike{}).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 减少点赞数
		if err := tx.Model(&article).Update("like_count", gorm.Expr("like_count - ?", 1)).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		// 添加点赞
		like := model.ArticleLike{
			ArticleID: articleID,
			UserID:    userID,
		}
		if err := tx.Create(&like).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 增加点赞数
		if err := tx.Model(&article).Update("like_count", gorm.Expr("like_count + ?", 1)).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// GetArticleCategories 获取文章分类统计
func GetArticleCategories() ([]model.CategoryStat, error) {
	var stats []model.CategoryStat

	if err := model.DB.Table("cms_categories").
		Select("cms_categories.category_id, cms_categories.category_name, COUNT(cms_articles.article_id) as article_count").
		Joins("LEFT JOIN cms_articles ON cms_categories.category_id = cms_articles.category_id AND cms_articles.is_published = true").
		Group("cms_categories.category_id").
		Order("article_count DESC").
		Find(&stats).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

// GetArticleTags 获取文章标签统计
func GetArticleTags() ([]model.TagStat, error) {
	var stats []model.TagStat

	if err := model.DB.Table("cms_tags").
		Select("cms_tags.tag_id, cms_tags.tag_name, cms_tags.tag_key, COUNT(cms_article_tags.article_id) as article_count").
		Joins("LEFT JOIN cms_article_tags ON cms_tags.tag_id = cms_article_tags.tag_id").
		Joins("LEFT JOIN cms_articles ON cms_article_tags.article_id = cms_articles.article_id AND cms_articles.is_published = true").
		Group("cms_tags.tag_id").
		Order("article_count DESC").
		Find(&stats).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

// GetArticleArchives 获取文章归档
func GetArticleArchives() ([]model.ArchiveStat, error) {
	var stats []model.ArchiveStat

	if err := model.DB.Table("cms_articles").
		Select("DATE_FORMAT(published_at, '%Y-%m') as month, COUNT(*) as article_count").
		Where("is_published = true").
		Group("month").
		Order("month DESC").
		Find(&stats).Error; err != nil {
		return nil, err
	}

	return stats, nil
}
