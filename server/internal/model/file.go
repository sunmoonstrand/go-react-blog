package model

import (
	"time"
)

// File 文件模型
type File struct {
	FileID       int64     `gorm:"column:file_id;primaryKey;autoIncrement" json:"file_id"`
	UserID       *int      `gorm:"column:user_id" json:"user_id"`
	OriginalName string    `gorm:"column:original_name;size:255;not null" json:"original_name"`
	FileName     string    `gorm:"column:file_name;size:100;not null" json:"file_name"`
	FilePath     string    `gorm:"column:file_path;size:255;not null" json:"file_path"`
	FileExt      string    `gorm:"column:file_ext;size:20;not null" json:"file_ext"`
	FileSize     int64     `gorm:"column:file_size;not null" json:"file_size"`
	MimeType     string    `gorm:"column:mime_type;size:100;not null" json:"mime_type"`
	StorageType  int8      `gorm:"column:storage_type;not null;default:1" json:"storage_type"`
	UseTimes     int       `gorm:"column:use_times;not null;default:0" json:"use_times"`
	IsPublic     bool      `gorm:"column:is_public;not null;default:true" json:"is_public"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	User         *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (File) TableName() string {
	return "sys_files"
}

// FileUploadForm 文件上传表单
type FileUploadForm struct {
	IsPublic bool `form:"is_public" json:"is_public" example:"true"`
}

// FileResponse 文件信息响应
type FileResponse struct {
	FileID       int64     `json:"file_id"`
	UserID       *int      `json:"user_id"`
	Username     string    `json:"username,omitempty"`
	OriginalName string    `json:"original_name"`
	FileName     string    `json:"file_name"`
	FilePath     string    `json:"file_path"`
	FileExt      string    `json:"file_ext"`
	FileSize     int64     `json:"file_size"`
	MimeType     string    `json:"mime_type"`
	StorageType  int8      `json:"storage_type"`
	UseTimes     int       `json:"use_times"`
	IsPublic     bool      `json:"is_public"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	URL          string    `json:"url"`
}

// FileQueryParams 文件查询参数
type FileQueryParams struct {
	UserID      *int   `form:"user_id" json:"user_id"`
	FileExt     string `form:"file_ext" json:"file_ext"`
	StorageType *int8  `form:"storage_type" json:"storage_type"`
	IsPublic    *bool  `form:"is_public" json:"is_public"`
	Keyword     string `form:"keyword" json:"keyword"`
	StartTime   string `form:"start_time" json:"start_time"`
	EndTime     string `form:"end_time" json:"end_time"`
	Page        int    `form:"page" json:"page" binding:"required,min=1" default:"1"`
	PageSize    int    `form:"page_size" json:"page_size" binding:"required,min=1,max=100" default:"10"`
}
