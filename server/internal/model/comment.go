package model

import (
	"time"
)

// Comment 评论模型
type Comment struct {
	CommentID    int64      `gorm:"column:comment_id;primaryKey;autoIncrement" json:"comment_id"`
	ArticleID    int64      `gorm:"column:article_id;not null" json:"article_id"`
	UserID       int        `gorm:"column:user_id;not null" json:"user_id"`
	ParentID     *int64     `gorm:"column:parent_id" json:"parent_id"`
	RootID       *int64     `gorm:"column:root_id" json:"root_id"`
	Content      string     `gorm:"column:content;not null" json:"content"`
	IPAddress    string     `gorm:"column:ip_address" json:"ip_address"`
	UserAgent    string     `gorm:"column:user_agent" json:"user_agent"`
	LikedCount   int        `gorm:"column:liked_count;not null;default:0" json:"liked_count"`
	IsApproved   bool       `gorm:"column:is_approved;not null;default:false" json:"is_approved"`
	IsAdminReply bool       `gorm:"column:is_admin_reply;not null;default:false" json:"is_admin_reply"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	User         User       `gorm:"foreignKey:UserID" json:"user"`
	Article      Article    `gorm:"foreignKey:ArticleID" json:"article"`
	Parent       *Comment   `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children     []*Comment `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

// TableName 指定表名
func (Comment) TableName() string {
	return "cms_comments"
}

// CommentCreateForm 评论创建表单
type CommentCreateForm struct {
	ArticleID int64  `json:"article_id" binding:"required" example:"1"`
	ParentID  *int64 `json:"parent_id" example:"0"`
	Content   string `json:"content" binding:"required" example:"这是一条评论内容"`
}

// CommentUpdateForm 评论更新表单
type CommentUpdateForm struct {
	Content    string `json:"content" binding:"required" example:"更新后的评论内容"`
	IsApproved bool   `json:"is_approved" example:"true"`
}

// CommentQueryParams 评论查询参数
type CommentQueryParams struct {
	ArticleID  int64  `form:"article_id" json:"article_id"`
	UserID     int    `form:"user_id" json:"user_id"`
	IsApproved *bool  `form:"is_approved" json:"is_approved"`
	Keyword    string `form:"keyword" json:"keyword"`
	StartTime  string `form:"start_time" json:"start_time"`
	EndTime    string `form:"end_time" json:"end_time"`
	Page       int    `form:"page" json:"page" binding:"required,min=1" default:"1"`
	PageSize   int    `form:"page_size" json:"page_size" binding:"required,min=1,max=100" default:"10"`
}

// CommentResponse 评论信息响应
type CommentResponse struct {
	CommentID    int64              `json:"comment_id"`
	ArticleID    int64              `json:"article_id"`
	ArticleTitle string             `json:"article_title"`
	UserID       int                `json:"user_id"`
	Username     string             `json:"username"`
	Nickname     string             `json:"nickname"`
	Avatar       string             `json:"avatar"`
	ParentID     *int64             `json:"parent_id"`
	RootID       *int64             `json:"root_id"`
	Content      string             `json:"content"`
	LikedCount   int                `json:"liked_count"`
	IsApproved   bool               `json:"is_approved"`
	IsAdminReply bool               `json:"is_admin_reply"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	Children     []*CommentResponse `json:"children,omitempty"`
}
