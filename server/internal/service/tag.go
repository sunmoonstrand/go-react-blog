package service

import (
	"errors"

	"github.com/sunmoonstrand/go-react-blog/server/internal/model"
	"gorm.io/gorm"
)

// CreateTag 创建标签
func CreateTag(form model.TagCreateForm) (int, error) {
	// 检查标签名是否已存在
	var count int64
	if err := model.DB.Model(&model.Tag{}).Where("tag_name = ?", form.TagName).Count(&count).Error; err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, errors.New("标签名已存在")
	}

	// 检查标签键是否已存在
	if err := model.DB.Model(&model.Tag{}).Where("tag_key = ?", form.TagKey).Count(&count).Error; err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, errors.New("标签键已存在")
	}

	// 创建标签
	tag := model.Tag{
		TagName: form.TagName,
		TagKey:  form.TagKey,
	}

	// 保存标签
	if err := model.DB.Create(&tag).Error; err != nil {
		return 0, err
	}

	return tag.TagID, nil
}

// UpdateTag 更新标签
func UpdateTag(tagID int, form model.TagUpdateForm) error {
	// 检查标签是否存在
	var tag model.Tag
	if err := model.DB.First(&tag, tagID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("标签不存在")
		}
		return err
	}

	// 检查标签名是否已被其他标签使用
	if form.TagName != "" && form.TagName != tag.TagName {
		var count int64
		if err := model.DB.Model(&model.Tag{}).
			Where("tag_name = ? AND tag_id != ?", form.TagName, tagID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("标签名已被其他标签使用")
		}
	}

	// 检查标签键是否已被其他标签使用
	if form.TagKey != "" && form.TagKey != tag.TagKey {
		var count int64
		if err := model.DB.Model(&model.Tag{}).
			Where("tag_key = ? AND tag_id != ?", form.TagKey, tagID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("标签键已被其他标签使用")
		}
	}

	// 更新标签
	updates := map[string]interface{}{}
	if form.TagName != "" {
		updates["tag_name"] = form.TagName
	}
	if form.TagKey != "" {
		updates["tag_key"] = form.TagKey
	}

	if err := model.DB.Model(&tag).Updates(updates).Error; err != nil {
		return err
	}

	return nil
}

// GetTagByID 根据ID获取标签
func GetTagByID(tagID int) (*model.Tag, error) {
	var tag model.Tag
	if err := model.DB.First(&tag, tagID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("标签不存在")
		}
		return nil, err
	}
	return &tag, nil
}

// ListTags 获取标签列表
func ListTags(params model.TagQueryParams) (*model.PageResult, error) {
	var tags []model.Tag
	var total int64

	// 构建查询
	query := model.DB.Model(&model.Tag{})

	// 应用过滤条件
	if params.TagName != "" {
		query = query.Where("tag_name LIKE ?", "%"+params.TagName+"%")
	}
	if params.TagKey != "" {
		query = query.Where("tag_key LIKE ?", "%"+params.TagKey+"%")
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	if err := query.Offset(offset).Limit(params.PageSize).Order("tag_id DESC").Find(&tags).Error; err != nil {
		return nil, err
	}

	// 转换为响应对象
	var tagResponses []model.TagResponse
	for _, tag := range tags {
		// 查询标签关联的文章数量
		var articleCount int64
		if err := model.DB.Model(&model.ArticleTag{}).Where("tag_id = ?", tag.TagID).Count(&articleCount).Error; err != nil {
			return nil, err
		}

		tagResponses = append(tagResponses, model.TagResponse{
			TagID:        tag.TagID,
			TagName:      tag.TagName,
			TagKey:       tag.TagKey,
			ArticleCount: int(articleCount),
			CreatedAt:    tag.CreatedAt,
			UpdatedAt:    tag.UpdatedAt,
		})
	}

	return model.NewPageResult(tagResponses, total, params.Page, params.PageSize), nil
}

// DeleteTag 删除标签
func DeleteTag(tagID int) error {
	// 检查标签是否存在
	var tag model.Tag
	if err := model.DB.First(&tag, tagID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("标签不存在")
		}
		return err
	}

	// 检查是否有关联的文章
	var count int64
	if err := model.DB.Model(&model.ArticleTag{}).Where("tag_id = ?", tagID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("该标签下有文章，不能删除")
	}

	// 删除标签
	if err := model.DB.Delete(&tag).Error; err != nil {
		return err
	}

	return nil
}

// GetAllTags 获取所有标签（用于下拉选择）
func GetAllTags() ([]model.Option, error) {
	var tags []model.Tag
	if err := model.DB.Order("tag_id DESC").Find(&tags).Error; err != nil {
		return nil, err
	}

	var options []model.Option
	for _, tag := range tags {
		options = append(options, model.Option{
			Label: tag.TagName,
			Value: tag.TagID,
		})
	}

	return options, nil
}

// GetTagsByArticleID 获取文章的标签
func GetTagsByArticleID(articleID int) ([]model.TagInfo, error) {
	var tags []model.Tag
	if err := model.DB.Table("cms_tags").
		Select("cms_tags.*").
		Joins("JOIN cms_article_tags ON cms_tags.tag_id = cms_article_tags.tag_id").
		Where("cms_article_tags.article_id = ?", articleID).
		Find(&tags).Error; err != nil {
		return nil, err
	}

	var tagInfos []model.TagInfo
	for _, tag := range tags {
		tagInfos = append(tagInfos, model.TagInfo{
			TagID:   tag.TagID,
			TagName: tag.TagName,
			TagKey:  tag.TagKey,
		})
	}

	return tagInfos, nil
}

// GetHotTags 获取热门标签
func GetHotTags(limit int) ([]model.TagStat, error) {
	var stats []model.TagStat

	if err := model.DB.Table("cms_tags").
		Select("cms_tags.tag_id, cms_tags.tag_name, cms_tags.tag_key, COUNT(cms_article_tags.article_id) as article_count").
		Joins("LEFT JOIN cms_article_tags ON cms_tags.tag_id = cms_article_tags.tag_id").
		Joins("LEFT JOIN cms_articles ON cms_article_tags.article_id = cms_articles.article_id AND cms_articles.is_published = true").
		Group("cms_tags.tag_id").
		Order("article_count DESC").
		Limit(limit).
		Find(&stats).Error; err != nil {
		return nil, err
	}

	return stats, nil
}
