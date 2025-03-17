package model

import (
	"time"
)

// Article 文章模型
type Article struct {
	ArticleID      int64          `gorm:"column:article_id;primaryKey;autoIncrement" json:"article_id"`
	UserID         int            `gorm:"column:user_id;not null" json:"user_id"`
	Title          string         `gorm:"column:title;size:200;not null" json:"title"`
	ArticleKey     string         `gorm:"column:article_key;size:200;not null;unique" json:"article_key"`
	Summary        string         `gorm:"column:summary;size:500" json:"summary"`
	Thumbnail      string         `gorm:"column:thumbnail;size:255" json:"thumbnail"`
	Status         int8           `gorm:"column:status;not null;default:1" json:"status"`
	ArticleType    int8           `gorm:"column:article_type;not null;default:1" json:"article_type"`
	ViewCount      int            `gorm:"column:view_count;not null;default:0" json:"view_count"`
	LikeCount      int            `gorm:"column:like_count;not null;default:0" json:"like_count"`
	CommentCount   int            `gorm:"column:comment_count;not null;default:0" json:"comment_count"`
	AllowComment   bool           `gorm:"column:allow_comment;not null;default:true" json:"allow_comment"`
	IsTop          bool           `gorm:"column:is_top;not null;default:false" json:"is_top"`
	IsRecommend    bool           `gorm:"column:is_recommend;not null;default:false" json:"is_recommend"`
	SEOTitle       string         `gorm:"column:seo_title;size:100" json:"seo_title"`
	SEOKeywords    string         `gorm:"column:seo_keywords;size:200" json:"seo_keywords"`
	SEODescription string         `gorm:"column:seo_description;size:300" json:"seo_description"`
	SourceURL      string         `gorm:"column:source_url;size:255" json:"source_url"`
	SourceName     string         `gorm:"column:source_name;size:100" json:"source_name"`
	PublishTime    time.Time      `gorm:"column:publish_time" json:"publish_time"`
	CreatedAt      time.Time      `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	User           User           `gorm:"foreignKey:UserID" json:"user"`
	Content        ArticleContent `gorm:"foreignKey:ArticleID" json:"content"`
	Categories     []Category     `gorm:"many2many:cms_article_categories;foreignKey:ArticleID;joinForeignKey:ArticleID;References:CategoryID;joinReferences:CategoryID" json:"categories"`
	Tags           []Tag          `gorm:"many2many:cms_article_tags;foreignKey:ArticleID;joinForeignKey:ArticleID;References:TagID;joinReferences:TagID" json:"tags"`
}

// TableName 指定表名
func (Article) TableName() string {
	return "cms_articles"
}

// ArticleContent 文章内容模型
type ArticleContent struct {
	ContentID     int64     `gorm:"column:content_id;primaryKey;autoIncrement" json:"content_id"`
	ArticleID     int64     `gorm:"column:article_id;not null" json:"article_id"`
	Content       string    `gorm:"column:content;not null" json:"content"`
	ContentFormat int8      `gorm:"column:content_format;not null;default:1" json:"content_format"`
	Version       int       `gorm:"column:version;not null;default:1" json:"version"`
	IsCurrent     bool      `gorm:"column:is_current;not null;default:true" json:"is_current"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName 指定表名
func (ArticleContent) TableName() string {
	return "cms_article_contents"
}

// ArticleCreateForm 文章创建表单
type ArticleCreateForm struct {
	Title          string `json:"title" binding:"required,max=200" example:"文章标题"`
	ArticleKey     string `json:"article_key" binding:"required,max=200" example:"article-key"`
	Content        string `json:"content" binding:"required" example:"文章内容..."`
	ContentFormat  int8   `json:"content_format" binding:"required,oneof=1 2" example:"1"`
	Summary        string `json:"summary" binding:"omitempty,max=500" example:"文章摘要..."`
	Thumbnail      string `json:"thumbnail" binding:"omitempty,max=255" example:"http://example.com/image.jpg"`
	Status         int8   `json:"status" binding:"required,oneof=1 2 3 4" example:"1"`
	ArticleType    int8   `json:"article_type" binding:"required,oneof=1 2 3 4" example:"1"`
	CategoryIDs    []int  `json:"category_ids" binding:"required" example:"1,2"`
	TagIDs         []int  `json:"tag_ids" binding:"omitempty" example:"1,2,3"`
	AllowComment   bool   `json:"allow_comment" example:"true"`
	IsTop          bool   `json:"is_top" example:"false"`
	IsRecommend    bool   `json:"is_recommend" example:"false"`
	SEOTitle       string `json:"seo_title" binding:"omitempty,max=100" example:"SEO标题"`
	SEOKeywords    string `json:"seo_keywords" binding:"omitempty,max=200" example:"关键词1,关键词2"`
	SEODescription string `json:"seo_description" binding:"omitempty,max=300" example:"SEO描述..."`
	SourceURL      string `json:"source_url" binding:"omitempty,max=255" example:"http://example.com/source"`
	SourceName     string `json:"source_name" binding:"omitempty,max=100" example:"来源网站"`
}

// ArticleUpdateForm 文章更新表单
type ArticleUpdateForm struct {
	Title          string `json:"title" binding:"omitempty,max=200" example:"文章标题"`
	ArticleKey     string `json:"article_key" binding:"omitempty,max=200" example:"article-key"`
	Content        string `json:"content" binding:"omitempty" example:"文章内容..."`
	ContentFormat  int8   `json:"content_format" binding:"omitempty,oneof=1 2" example:"1"`
	Summary        string `json:"summary" binding:"omitempty,max=500" example:"文章摘要..."`
	Thumbnail      string `json:"thumbnail" binding:"omitempty,max=255" example:"http://example.com/image.jpg"`
	Status         int8   `json:"status" binding:"omitempty,oneof=1 2 3 4" example:"1"`
	ArticleType    int8   `json:"article_type" binding:"omitempty,oneof=1 2 3 4" example:"1"`
	CategoryIDs    []int  `json:"category_ids" binding:"omitempty" example:"1,2"`
	TagIDs         []int  `json:"tag_ids" binding:"omitempty" example:"1,2,3"`
	AllowComment   bool   `json:"allow_comment" example:"true"`
	IsTop          bool   `json:"is_top" example:"false"`
	IsRecommend    bool   `json:"is_recommend" example:"false"`
	SEOTitle       string `json:"seo_title" binding:"omitempty,max=100" example:"SEO标题"`
	SEOKeywords    string `json:"seo_keywords" binding:"omitempty,max=200" example:"关键词1,关键词2"`
	SEODescription string `json:"seo_description" binding:"omitempty,max=300" example:"SEO描述..."`
	SourceURL      string `json:"source_url" binding:"omitempty,max=255" example:"http://example.com/source"`
	SourceName     string `json:"source_name" binding:"omitempty,max=100" example:"来源网站"`
}

// ArticleQueryParams 文章查询参数
type ArticleQueryParams struct {
	Keyword     string `form:"keyword" json:"keyword"`
	Status      int8   `form:"status" json:"status"`
	CategoryID  int    `form:"category_id" json:"category_id"`
	TagID       int    `form:"tag_id" json:"tag_id"`
	UserID      int    `form:"user_id" json:"user_id"`
	ArticleType int8   `form:"article_type" json:"article_type"`
	IsTop       *bool  `form:"is_top" json:"is_top"`
	IsRecommend *bool  `form:"is_recommend" json:"is_recommend"`
	StartTime   string `form:"start_time" json:"start_time"`
	EndTime     string `form:"end_time" json:"end_time"`
	OrderBy     string `form:"order_by" json:"order_by"`
	Page        int    `form:"page" json:"page" binding:"required,min=1" default:"1"`
	PageSize    int    `form:"page_size" json:"page_size" binding:"required,min=1,max=100" default:"10"`
}

// ArticleResponse 文章信息响应
type ArticleResponse struct {
	ArticleID       int64     `json:"article_id"`
	UserID          int       `json:"user_id"`
	Author          string    `json:"author"`
	Title           string    `json:"title"`
	ArticleKey      string    `json:"article_key"`
	Summary         string    `json:"summary"`
	Thumbnail       string    `json:"thumbnail"`
	Status          int8      `json:"status"`
	StatusName      string    `json:"status_name"`
	ArticleType     int8      `json:"article_type"`
	ArticleTypeName string    `json:"article_type_name"`
	ViewCount       int       `json:"view_count"`
	LikeCount       int       `json:"like_count"`
	CommentCount    int       `json:"comment_count"`
	AllowComment    bool      `json:"allow_comment"`
	IsTop           bool      `json:"is_top"`
	IsRecommend     bool      `json:"is_recommend"`
	PublishTime     time.Time `json:"publish_time"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Categories      []string  `json:"categories"`
	Tags            []string  `json:"tags"`
}

// ArticleDetailResponse 文章详情响应
type ArticleDetailResponse struct {
	ArticleResponse
	Content        string `json:"content"`
	ContentFormat  int8   `json:"content_format"`
	SEOTitle       string `json:"seo_title"`
	SEOKeywords    string `json:"seo_keywords"`
	SEODescription string `json:"seo_description"`
	SourceURL      string `json:"source_url"`
	SourceName     string `json:"source_name"`
}
