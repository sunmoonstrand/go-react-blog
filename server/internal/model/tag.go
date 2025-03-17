package model

import (
	"time"
)

// Tag 标签模型
type Tag struct {
	TagID        int       `gorm:"column:tag_id;primaryKey;autoIncrement" json:"tag_id"`
	TagName      string    `gorm:"column:tag_name;size:50;not null;unique" json:"tag_name"`
	TagKey       string    `gorm:"column:tag_key;size:50;not null;unique" json:"tag_key"`
	Description  string    `gorm:"column:description" json:"description"`
	Thumbnail    string    `gorm:"column:thumbnail;size:255" json:"thumbnail"`
	SortOrder    int16     `gorm:"column:sort_order;not null;default:0" json:"sort_order"`
	IsVisible    bool      `gorm:"column:is_visible;not null;default:true" json:"is_visible"`
	ArticleCount int       `gorm:"column:article_count;not null;default:0" json:"article_count"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	Articles     []Article `gorm:"many2many:cms_article_tags;foreignKey:TagID;joinForeignKey:TagID;References:ArticleID;joinReferences:ArticleID" json:"articles,omitempty"`
}

// TableName 指定表名
func (Tag) TableName() string {
	return "cms_tags"
}

// TagCreateForm 标签创建表单
type TagCreateForm struct {
	TagName     string `json:"tag_name" binding:"required,max=50" example:"Go语言"`
	TagKey      string `json:"tag_key" binding:"required,max=50" example:"golang"`
	Description string `json:"description" binding:"omitempty" example:"Go语言相关文章标签"`
	Thumbnail   string `json:"thumbnail" binding:"omitempty,max=255" example:"http://example.com/image.jpg"`
	SortOrder   int16  `json:"sort_order" binding:"omitempty" example:"1"`
	IsVisible   bool   `json:"is_visible" example:"true"`
}

// TagUpdateForm 标签更新表单
type TagUpdateForm struct {
	TagName     string `json:"tag_name" binding:"omitempty,max=50" example:"Go语言"`
	TagKey      string `json:"tag_key" binding:"omitempty,max=50" example:"golang"`
	Description string `json:"description" binding:"omitempty" example:"Go语言相关文章标签"`
	Thumbnail   string `json:"thumbnail" binding:"omitempty,max=255" example:"http://example.com/image.jpg"`
	SortOrder   int16  `json:"sort_order" binding:"omitempty" example:"1"`
	IsVisible   bool   `json:"is_visible" example:"true"`
}

// TagResponse 标签信息响应
type TagResponse struct {
	TagID        int       `json:"tag_id"`
	TagName      string    `json:"tag_name"`
	TagKey       string    `json:"tag_key"`
	Description  string    `json:"description"`
	Thumbnail    string    `json:"thumbnail"`
	SortOrder    int16     `json:"sort_order"`
	IsVisible    bool      `json:"is_visible"`
	ArticleCount int       `json:"article_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
