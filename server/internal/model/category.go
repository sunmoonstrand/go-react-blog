package model

import (
	"time"
)

// Category 分类模型
type Category struct {
	CategoryID     int         `gorm:"column:category_id;primaryKey;autoIncrement" json:"category_id"`
	ParentID       *int        `gorm:"column:parent_id" json:"parent_id"`
	CategoryName   string      `gorm:"column:category_name;size:50;not null" json:"category_name"`
	CategoryKey    string      `gorm:"column:category_key;size:50;not null;unique" json:"category_key"`
	Path           string      `gorm:"column:path;not null" json:"path"`
	Description    string      `gorm:"column:description" json:"description"`
	Thumbnail      string      `gorm:"column:thumbnail;size:255" json:"thumbnail"`
	Icon           string      `gorm:"column:icon;size:100" json:"icon"`
	SortOrder      int16       `gorm:"column:sort_order;not null;default:0" json:"sort_order"`
	IsVisible      bool        `gorm:"column:is_visible;not null;default:true" json:"is_visible"`
	SEOTitle       string      `gorm:"column:seo_title;size:100" json:"seo_title"`
	SEOKeywords    string      `gorm:"column:seo_keywords;size:200" json:"seo_keywords"`
	SEODescription string      `gorm:"column:seo_description;size:300" json:"seo_description"`
	ArticleCount   int         `gorm:"column:article_count;not null;default:0" json:"article_count"`
	CreatedAt      time.Time   `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time   `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	Parent         *Category   `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children       []*Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Articles       []Article   `gorm:"many2many:cms_article_categories;foreignKey:CategoryID;joinForeignKey:CategoryID;References:ArticleID;joinReferences:ArticleID" json:"articles,omitempty"`
}

// TableName 指定表名
func (Category) TableName() string {
	return "cms_categories"
}

// CategoryCreateForm 分类创建表单
type CategoryCreateForm struct {
	ParentID       *int   `json:"parent_id" example:"0"`
	CategoryName   string `json:"category_name" binding:"required,max=50" example:"技术文章"`
	CategoryKey    string `json:"category_key" binding:"required,max=50" example:"tech"`
	Description    string `json:"description" binding:"omitempty" example:"技术相关文章分类"`
	Thumbnail      string `json:"thumbnail" binding:"omitempty,max=255" example:"http://example.com/image.jpg"`
	Icon           string `json:"icon" binding:"omitempty,max=100" example:"code"`
	SortOrder      int16  `json:"sort_order" binding:"omitempty" example:"1"`
	IsVisible      bool   `json:"is_visible" example:"true"`
	SEOTitle       string `json:"seo_title" binding:"omitempty,max=100" example:"技术文章分类"`
	SEOKeywords    string `json:"seo_keywords" binding:"omitempty,max=200" example:"技术,编程,开发"`
	SEODescription string `json:"seo_description" binding:"omitempty,max=300" example:"包含各种技术相关文章的分类"`
}

// CategoryUpdateForm 分类更新表单
type CategoryUpdateForm struct {
	ParentID       *int   `json:"parent_id" example:"0"`
	CategoryName   string `json:"category_name" binding:"omitempty,max=50" example:"技术文章"`
	CategoryKey    string `json:"category_key" binding:"omitempty,max=50" example:"tech"`
	Description    string `json:"description" binding:"omitempty" example:"技术相关文章分类"`
	Thumbnail      string `json:"thumbnail" binding:"omitempty,max=255" example:"http://example.com/image.jpg"`
	Icon           string `json:"icon" binding:"omitempty,max=100" example:"code"`
	SortOrder      int16  `json:"sort_order" binding:"omitempty" example:"1"`
	IsVisible      bool   `json:"is_visible" example:"true"`
	SEOTitle       string `json:"seo_title" binding:"omitempty,max=100" example:"技术文章分类"`
	SEOKeywords    string `json:"seo_keywords" binding:"omitempty,max=200" example:"技术,编程,开发"`
	SEODescription string `json:"seo_description" binding:"omitempty,max=300" example:"包含各种技术相关文章的分类"`
}

// CategoryResponse 分类信息响应
type CategoryResponse struct {
	CategoryID   int                 `json:"category_id"`
	ParentID     *int                `json:"parent_id"`
	CategoryName string              `json:"category_name"`
	CategoryKey  string              `json:"category_key"`
	Description  string              `json:"description"`
	Thumbnail    string              `json:"thumbnail"`
	Icon         string              `json:"icon"`
	SortOrder    int16               `json:"sort_order"`
	IsVisible    bool                `json:"is_visible"`
	ArticleCount int                 `json:"article_count"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
	Children     []*CategoryResponse `json:"children,omitempty"`
}

// CategoryDetailResponse 分类详情响应
type CategoryDetailResponse struct {
	CategoryResponse
	SEOTitle       string `json:"seo_title"`
	SEOKeywords    string `json:"seo_keywords"`
	SEODescription string `json:"seo_description"`
	Path           string `json:"path"`
}
