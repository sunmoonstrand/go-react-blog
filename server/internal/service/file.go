package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sunmoonstrand/go-react-blog/server/config"
	"github.com/sunmoonstrand/go-react-blog/server/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UploadFile 上传文件
func UploadFile(file *multipart.FileHeader, userID int, isPublic bool) (*model.UploadResult, error) {
	// 获取文件信息
	originalName := file.Filename
	fileExt := strings.ToLower(filepath.Ext(originalName))
	fileSize := file.Size
	mimeType := file.Header.Get("Content-Type")

	// 生成唯一文件名
	fileName := fmt.Sprintf("%s%s", uuid.New().String(), fileExt)

	// 确定存储路径
	uploadDir := config.AppConfig.Upload.Path
	if uploadDir == "" {
		uploadDir = "uploads"
	}

	// 按日期分目录
	datePath := time.Now().Format("2006/01/02")
	uploadPath := filepath.Join(uploadDir, datePath)

	// 创建目录
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		zap.L().Error("创建上传目录失败",
			zap.String("path", uploadPath),
			zap.Error(err),
		)
		return nil, errors.New("创建上传目录失败")
	}

	// 完整文件路径
	filePath := filepath.Join(uploadPath, fileName)

	// 打开源文件
	src, err := file.Open()
	if err != nil {
		zap.L().Error("打开上传文件失败",
			zap.String("filename", originalName),
			zap.Error(err),
		)
		return nil, errors.New("打开上传文件失败")
	}
	defer src.Close()

	// 创建目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		zap.L().Error("创建目标文件失败",
			zap.String("path", filePath),
			zap.Error(err),
		)
		return nil, errors.New("创建目标文件失败")
	}
	defer dst.Close()

	// 复制文件内容
	if _, err = io.Copy(dst, src); err != nil {
		zap.L().Error("复制文件内容失败",
			zap.String("src", originalName),
			zap.String("dst", filePath),
			zap.Error(err),
		)
		return nil, errors.New("复制文件内容失败")
	}

	// 相对路径，用于存储到数据库
	relativePath := filepath.Join(datePath, fileName)

	// 存储到数据库
	fileRecord := model.File{
		UserID:       userID,
		OriginalName: originalName,
		FileName:     fileName,
		FilePath:     relativePath,
		FileExt:      fileExt,
		FileSize:     int(fileSize),
		MimeType:     mimeType,
		StorageType:  1, // 本地存储
		UseTimes:     0,
		IsPublic:     isPublic,
	}

	if err := model.DB.Create(&fileRecord).Error; err != nil {
		zap.L().Error("保存文件记录失败",
			zap.String("filename", originalName),
			zap.Error(err),
		)
		// 删除已上传的文件
		os.Remove(filePath)
		return nil, errors.New("保存文件记录失败")
	}

	// 生成访问URL
	baseURL := config.AppConfig.Upload.BaseURL
	if baseURL == "" {
		baseURL = "/uploads"
	}
	fileURL := fmt.Sprintf("%s/%s", baseURL, relativePath)

	// 返回结果
	result := &model.UploadResult{
		FileID:       fileRecord.FileID,
		OriginalName: originalName,
		FileName:     fileName,
		FileExt:      fileExt,
		FileSize:     int(fileSize),
		URL:          fileURL,
	}

	return result, nil
}

// GetFileByID 根据ID获取文件
func GetFileByID(fileID int) (*model.File, error) {
	var file model.File
	if err := model.DB.First(&file, fileID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("文件不存在")
		}
		return nil, err
	}
	return &file, nil
}

// DeleteFile 删除文件
func DeleteFile(fileID int, userID int) error {
	// 检查文件是否存在
	var file model.File
	if err := model.DB.First(&file, fileID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("文件不存在")
		}
		return err
	}

	// 检查是否有权限删除（只有上传者或管理员可以删除）
	if file.UserID != userID {
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
			return errors.New("无权限删除该文件")
		}
	}

	// 检查文件是否被使用
	if file.UseTimes > 0 {
		return errors.New("文件正在被使用，不能删除")
	}

	// 删除物理文件
	uploadDir := config.AppConfig.Upload.Path
	if uploadDir == "" {
		uploadDir = "uploads"
	}
	filePath := filepath.Join(uploadDir, file.FilePath)

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		zap.L().Error("删除物理文件失败",
			zap.String("path", filePath),
			zap.Error(err),
		)
		// 不返回错误，继续删除数据库记录
	}

	// 删除数据库记录
	if err := model.DB.Delete(&file).Error; err != nil {
		return err
	}

	return nil
}

// ListFiles 获取文件列表
func ListFiles(params model.FileQueryParams) (*model.PageResult, error) {
	var files []model.File
	var total int64

	// 构建查询
	query := model.DB.Model(&model.File{})

	// 应用过滤条件
	if params.UserID != nil {
		query = query.Where("user_id = ?", *params.UserID)
	}
	if params.FileExt != "" {
		query = query.Where("file_ext = ?", params.FileExt)
	}
	if params.StorageType != nil {
		query = query.Where("storage_type = ?", *params.StorageType)
	}
	if params.IsPublic != nil {
		query = query.Where("is_public = ?", *params.IsPublic)
	}
	if params.Keyword != "" {
		query = query.Where("original_name LIKE ? OR file_name LIKE ?",
			"%"+params.Keyword+"%", "%"+params.Keyword+"%")
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
	if err := query.Offset(offset).Limit(params.PageSize).Order("file_id DESC").Find(&files).Error; err != nil {
		return nil, err
	}

	// 转换为响应对象
	var fileResponses []model.FileResponse
	baseURL := config.AppConfig.Upload.BaseURL
	if baseURL == "" {
		baseURL = "/uploads"
	}

	for _, file := range files {
		fileURL := fmt.Sprintf("%s/%s", baseURL, file.FilePath)
		fileResponses = append(fileResponses, model.FileResponse{
			FileID:       file.FileID,
			UserID:       file.UserID,
			OriginalName: file.OriginalName,
			FileName:     file.FileName,
			FilePath:     file.FilePath,
			FileExt:      file.FileExt,
			FileSize:     file.FileSize,
			MimeType:     file.MimeType,
			StorageType:  file.StorageType,
			UseTimes:     file.UseTimes,
			IsPublic:     file.IsPublic,
			URL:          fileURL,
			CreatedAt:    file.CreatedAt,
			UpdatedAt:    file.UpdatedAt,
		})
	}

	return model.NewPageResult(fileResponses, total, params.Page, params.PageSize), nil
}

// UpdateFileUsage 更新文件使用次数
func UpdateFileUsage(fileID int, increment bool) error {
	// 检查文件是否存在
	var file model.File
	if err := model.DB.First(&file, fileID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("文件不存在")
		}
		return err
	}

	// 更新使用次数
	var expr string
	if increment {
		expr = "use_times + 1"
	} else {
		expr = "GREATEST(use_times - 1, 0)" // 确保不会小于0
	}

	if err := model.DB.Model(&file).Update("use_times", gorm.Expr(expr)).Error; err != nil {
		return err
	}

	return nil
}

// GetFileURL 获取文件URL
func GetFileURL(fileID int) (string, error) {
	var file model.File
	if err := model.DB.First(&file, fileID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("文件不存在")
		}
		return "", err
	}

	// 检查文件是否公开
	if !file.IsPublic {
		return "", errors.New("文件不公开")
	}

	// 生成访问URL
	baseURL := config.AppConfig.Upload.BaseURL
	if baseURL == "" {
		baseURL = "/uploads"
	}
	fileURL := fmt.Sprintf("%s/%s", baseURL, file.FilePath)

	return fileURL, nil
}
