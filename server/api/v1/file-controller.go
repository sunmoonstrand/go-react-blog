package v1

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/yourusername/blog/server/internal/logger"
	"github.com/yourusername/blog/server/internal/middleware"
	"github.com/yourusername/blog/server/internal/model"
	"github.com/yourusername/blog/server/internal/service/storage"
	"github.com/yourusername/blog/server/internal/utils/resp"
)

// FileController 文件控制器
type FileController struct {
	fileModel      *model.FileModel
	configModel    *model.ConfigModel
	storageService *storage.StorageService
}

// NewFileController 创建文件控制器实例
func NewFileController(fileModel *model.FileModel, configModel *model.ConfigModel, storageService *storage.StorageService) *FileController {
	return &FileController{
		fileModel:      fileModel,
		configModel:    configModel,
		storageService: storageService,
	}
}

// UploadFileResponse 上传文件响应结构
type UploadFileResponse struct {
	FileID   uint   `json:"file_id"`
	FileName string `json:"file_name"`
	FileType string `json:"file_type"`
	FileSize int64  `json:"file_size"`
	FileURL  string `json:"file_url"`
}

// GetFileListResponse 文件列表响应结构
type GetFileListResponse struct {
	FileID      uint      `json:"file_id"`
	FileName    string    `json:"file_name"`
	FileType    string    `json:"file_type"`
	FileSize    int64     `json:"file_size"`
	FileURL     string    `json:"file_url"`
	UploadUser  string    `json:"upload_user"`
	Description string    `json:"description"`
	UploadTime  time.Time `json:"upload_time"`
}

// UploadFile 上传文件
// @Summary 上传文件
// @Description 上传文件到服务器
// @Tags 文件管理
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param file formData file true "要上传的文件"
// @Param type formData string false "文件类型:image,document,media,other"
// @Param description formData string false "文件描述"
// @Success 200 {object} resp.Response{data=UploadFileResponse} "上传成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/files/upload [post]
func (fc *FileController) UploadFile(c *gin.Context) {
	// 获取上传配置
	uploadConfig, err := fc.configModel.GetUploadConfig()
	if err != nil {
		logger.Error("获取上传配置失败", "error", err)
		resp.FailWithMsg(c, "文件上传失败，请稍后重试")
		return
	}

	// 获取文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		resp.FailWithMsg(c, "未找到上传文件")
		return
	}
	defer file.Close()

	// 检查文件大小
	if header.Size > uploadConfig.MaxSize*1024*1024 {
		resp.FailWithMsg(c, fmt.Sprintf("文件大小超出限制，最大允许 %dMB", uploadConfig.MaxSize))
		return
	}

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		resp.FailWithMsg(c, "文件缺少扩展名")
		return
	}

	// 去除扩展名前面的点号
	ext = ext[1:]

	// 检查文件类型是否允许
	allowed := false
	for _, format := range uploadConfig.AllowedFormats {
		if ext == format {
			allowed = true
			break
		}
	}
	if !allowed {
		resp.FailWithMsg(c, fmt.Sprintf("不支持的文件类型，允许的类型: %s", strings.Join(uploadConfig.AllowedFormats, ", ")))
		return
	}

	// 生成文件名
	filename := uuid.New().String() + "." + ext

	// 获取用户自定义文件类别
	fileType := c.DefaultPostForm("type", "other")
	if fileType != "image" && fileType != "document" && fileType != "media" && fileType != "other" {
		fileType = "other"
	}

	// 获取文件描述
	description := c.DefaultPostForm("description", "")

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		resp.FailWithCode(c, http.StatusUnauthorized, "未授权")
		return
	}

	// 构建保存路径
	savePath := filepath.Join(
		uploadConfig.SavePath,
		fileType,
		time.Now().Format("2006/01/02"),
	)

	// 上传文件
	fileUrl, err := fc.storageService.UploadFile(file, filename, savePath, uploadConfig.Provider)
	if err != nil {
		logger.Error("上传文件失败", "filename", filename, "error", err)
		resp.FailWithMsg(c, "文件上传失败: "+err.Error())
		return
	}

	// 保存文件记录到数据库
	fileRecord := &model.File{
		UserID:      userID.(uint),
		FileName:    header.Filename,
		StoragePath: filepath.Join(savePath, filename),
		FileURL:     fileUrl,
		FileSize:    header.Size,
		FileType:    fileType,
		MimeType:    c.GetHeader("Content-Type"),
		Extension:   ext,
		Description: description,
	}

	if err := fc.fileModel.CreateFile(fileRecord); err != nil {
		logger.Error("保存文件记录失败", "filename", filename, "error", err)
		resp.FailWithMsg(c, "文件上传失败，请稍后重试")
		return
	}

	// 返回响应
	resp.OkWithData(c, UploadFileResponse{
		FileID:   fileRecord.ID,
		FileName: fileRecord.FileName,
		FileType: fileRecord.FileType,
		FileSize: fileRecord.FileSize,
		FileURL:  fileRecord.FileURL,
	})
}

// UploadImage 上传图片
// @Summary 上传图片
// @Description 专门用于上传图片的接口
// @Tags 文件管理
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param file formData file true "要上传的图片"
// @Param description formData string false "图片描述"
// @Success 200 {object} resp.Response{data=UploadFileResponse} "上传成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/files/upload/image [post]
func (fc *FileController) UploadImage(c *gin.Context) {
	// 获取上传配置
	uploadConfig, err := fc.configModel.GetUploadConfig()
	if err != nil {
		logger.Error("获取上传配置失败", "error", err)
		resp.FailWithMsg(c, "图片上传失败，请稍后重试")
		return
	}

	// 获取文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		resp.FailWithMsg(c, "未找到上传图片")
		return
	}
	defer file.Close()

	// 检查文件大小
	if header.Size > uploadConfig.MaxSize*1024*1024 {
		resp.FailWithMsg(c, fmt.Sprintf("图片大小超出限制，最大允许 %dMB", uploadConfig.MaxSize))
		return
	}

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		resp.FailWithMsg(c, "图片缺少扩展名")
		return
	}

	// 去除扩展名前面的点号
	ext = ext[1:]

	// 检查是否为图片类型
	imageFormats := []string{"jpg", "jpeg", "png", "gif", "bmp", "webp", "svg"}
	allowed := false
	for _, format := range imageFormats {
		if ext == format {
			allowed = true
			break
		}
	}
	if !allowed {
		resp.FailWithMsg(c, "不支持的图片类型，允许的类型: jpg, jpeg, png, gif, bmp, webp, svg")
		return
	}

	// 生成文件名
	filename := uuid.New().String() + "." + ext

	// 获取图片描述
	description := c.DefaultPostForm("description", "")

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		resp.FailWithCode(c, http.StatusUnauthorized, "未授权")
		return
	}

	// 构建保存路径
	savePath := filepath.Join(
		uploadConfig.SavePath,
		"image",
		time.Now().Format("2006/01/02"),
	)

	// 上传文件
	fileUrl, err := fc.storageService.UploadFile(file, filename, savePath, uploadConfig.Provider)
	if err != nil {
		logger.Error("上传图片失败", "filename", filename, "error", err)
		resp.FailWithMsg(c, "图片上传失败: "+err.Error())
		return
	}

	// 保存文件记录到数据库
	fileRecord := &model.File{
		UserID:      userID.(uint),
		FileName:    header.Filename,
		StoragePath: filepath.Join(savePath, filename),
		FileURL:     fileUrl,
		FileSize:    header.Size,
		FileType:    "image",
		MimeType:    c.GetHeader("Content-Type"),
		Extension:   ext,
		Description: description,
	}

	if err := fc.fileModel.CreateFile(fileRecord); err != nil {
		logger.Error("保存图片记录失败", "filename", filename, "error", err)
		resp.FailWithMsg(c, "图片上传失败，请稍后重试")
		return
	}

	// 返回响应
	resp.OkWithData(c, UploadFileResponse{
		FileID:   fileRecord.ID,
		FileName: fileRecord.FileName,
		FileType: fileRecord.FileType,
		FileSize: fileRecord.FileSize,
		FileURL:  fileRecord.FileURL,
	})
}

// GetFileList 获取文件列表
// @Summary 获取文件列表
// @Description 分页获取上传的文件列表
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码，默认1" default(1)
// @Param page_size query int false "每页记录数，默认20" default(20)
// @Param type query string false "文件类型:image,document,media,other"
// @Param keyword query string false "搜索关键词"
// @Success 200 {object} resp.Response{data=resp.PageResult} "返回文件列表"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/files [get]
func (fc *FileController) GetFileList(c *gin.Context) {
	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	fileType := c.Query("type")
	keyword := c.Query("keyword")

	// 调用服务获取文件列表
	files, total, err := fc.fileModel.GetFileList(page, pageSize, fileType, keyword)
	if err != nil {
		logger.Error("获取文件列表失败", "error", err)
		resp.FailWithMsg(c, "获取文件列表失败")
		return
	}

	// 构建响应数据
	var respData []GetFileListResponse
	for _, file := range files {
		respData = append(respData, GetFileListResponse{
			FileID:      file.ID,
			FileName:    file.FileName,
			FileType:    file.FileType,
			FileSize:    file.FileSize,
			FileURL:     file.FileURL,
			UploadUser:  file.UserName,
			Description: file.Description,
			UploadTime:  file.CreatedAt,
		})
	}

	resp.OkWithPage(c, respData, total, page, pageSize)
}

// GetFile 获取文件详情
// @Summary 获取文件详情
// @Description 根据文件ID获取详细信息
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "文件ID"
// @Success 200 {object} resp.Response "返回文件详情"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 404 {object} resp.Response "文件不存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/files/{id} [get]
func (fc *FileController) GetFile(c *gin.Context) {
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil {
		resp.FailWithMsg(c, "无效的文件ID")
		return
	}

	// 获取文件详情
	file, err := fc.fileModel.GetFileByID(uint(fileID))
	if err != nil {
		if err.Error() == "record not found" {
			resp.FailWithCode(c, http.StatusNotFound, "文件不存在")
		} else {
			logger.Error("获取文件详情失败", "file_id", fileID, "error", err)
			resp.FailWithMsg(c, "获取文件详情失败")
		}
		return
	}

	// 构建响应数据
	resp.OkWithData(c, gin.H{
		"file_id":     file.ID,
		"file_name":   file.FileName,
		"file_type":   file.FileType,
		"file_size":   file.FileSize,
		"file_url":    file.FileURL,
		"mime_type":   file.MimeType,
		"extension":   file.Extension,
		"user_id":     file.UserID,
		"description": file.Description,
		"created_at":  file.CreatedAt,
		"updated_at":  file.UpdatedAt,
	})
}

// UpdateFile 更新文件信息
// @Summary 更新文件信息
// @Description 更新文件的描述等信息
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "文件ID"
// @Param data body struct{description string} true "文件描述"
// @Success 200 {object} resp.Response "更新成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 404 {object} resp.Response "文件不存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/files/{id} [put]
func (fc *FileController) UpdateFile(c *gin.Context) {
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil {
		resp.FailWithMsg(c, "无效的文件ID")
		return
	}

	// 获取文件详情
	file, err := fc.fileModel.GetFileByID(uint(fileID))
	if err != nil {
		if err.Error() == "record not found" {
			resp.FailWithCode(c, http.StatusNotFound, "文件不存在")
		} else {
			logger.Error("获取文件详情失败", "file_id", fileID, "error", err)
			resp.FailWithMsg(c, "更新文件失败")
		}
		return
	}

	// 解析请求数据
	var data struct {
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.FailWithValidation(c, err)
		return
	}

	// 更新文件信息
	if err := fc.fileModel.UpdateFile(uint(fileID), data.Description); err != nil {
		logger.Error("更新文件失败", "file_id", fileID, "error", err)
		resp.FailWithMsg(c, "更新文件失败")
		return
	}

	resp.OkWithMsg(c, "更新文件成功")
}

// DeleteFile 删除文件
// @Summary 删除文件
// @Description 根据ID删除文件
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "文件ID"
// @Success 200 {object} resp.Response "删除成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 404 {object} resp.Response "文件不存在"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/files/{id} [delete]
func (fc *FileController) DeleteFile(c *gin.Context) {
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil {
		resp.FailWithMsg(c, "无效的文件ID")
		return
	}

	// 获取上传配置
	uploadConfig, err := fc.configModel.GetUploadConfig()
	if err != nil {
		logger.Error("获取上传配置失败", "error", err)
		resp.FailWithMsg(c, "删除文件失败，请稍后重试")
		return
	}

	// 获取文件详情
	file, err := fc.fileModel.GetFileByID(uint(fileID))
	if err != nil {
		if err.Error() == "record not found" {
			resp.FailWithCode(c, http.StatusNotFound, "文件不存在")
		} else {
			logger.Error("获取文件详情失败", "file_id", fileID, "error", err)
			resp.FailWithMsg(c, "删除文件失败")
		}
		return
	}

	// 从存储服务删除文件
	if err := fc.storageService.DeleteFile(file.StoragePath, uploadConfig.Provider); err != nil {
		logger.Error("从存储服务删除文件失败", "file_id", fileID, "path", file.StoragePath, "error", err)
		// 继续执行，确保数据库记录被删除
	}

	// 从数据库删除文件记录
	if err := fc.fileModel.DeleteFile(uint(fileID)); err != nil {
		logger.Error("从数据库删除文件记录失败", "file_id", fileID, "error", err)
		resp.FailWithMsg(c, "删除文件失败")
		return
	}

	resp.OkWithMsg(c, "删除文件成功")
}

// BatchDeleteFiles 批量删除文件
// @Summary 批量删除文件
// @Description 批量删除多个文件
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param ids body []int true "文件ID数组"
// @Success 200 {object} resp.Response "删除成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/files/batch [delete]
func (fc *FileController) BatchDeleteFiles(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		resp.FailWithMsg(c, "无效的参数")
		return
	}

	if len(ids) == 0 {
		resp.FailWithMsg(c, "请选择要删除的文件")
		return
	}

	// 获取上传配置
	uploadConfig, err := fc.configModel.GetUploadConfig()
	if err != nil {
		logger.Error("获取上传配置失败", "error", err)
		resp.FailWithMsg(c, "批量删除文件失败，请稍后重试")
		return
	}

	// 获取文件列表
	files, err := fc.fileModel.GetFilesByIDs(ids)
	if err != nil {
		logger.Error("获取文件列表失败", "ids", ids, "error", err)
		resp.FailWithMsg(c, "批量删除文件失败")
		return
	}

	// 从存储服务批量删除文件
	for _, file := range files {
		if err := fc.storageService.DeleteFile(file.StoragePath, uploadConfig.Provider); err != nil {
			logger.Error("从存储服务删除文件失败", "file_id", file.ID, "path", file.StoragePath, "error", err)
			// 继续执行，确保尽可能多的文件被删除
		}
	}

	// 从数据库批量删除文件记录
	result, err := fc.fileModel.BatchDeleteFiles(ids)
	if err != nil {
		logger.Error("从数据库批量删除文件记录失败", "ids", ids, "error", err)
		resp.FailWithMsg(c, "批量删除文件失败")
		return
	}

	resp.OkWithData(c, gin.H{
		"success": result.Success,
		"failed":  result.Failed,
		"message": "批量删除文件完成",
	})
}

// RegisterRoutes 注册路由
func (fc *FileController) RegisterRoutes(router *gin.RouterGroup) {
	fileGroup := router.Group("/files")
	{
		// 获取文件列表
		fileGroup.GET("", middleware.RequirePermission("content:file:list"), fc.GetFileList)

		// 获取文件详情
		fileGroup.GET("/:id", middleware.RequirePermission("content:file:info"), fc.GetFile)

		// 更新文件信息
		fileGroup.PUT("/:id", middleware.RequirePermission("content:file:edit"), fc.UpdateFile)

		// 删除文件
		fileGroup.DELETE("/:id", middleware.RequirePermission("content:file:delete"), fc.DeleteFile)

		// 批量删除文件
		fileGroup.DELETE("/batch", middleware.RequirePermission("content:file:delete"), fc.BatchDeleteFiles)

		// 上传文件
		fileGroup.POST("/upload", middleware.RequirePermission("content:file:upload"), fc.UploadFile)

		// 上传图片
		fileGroup.POST("/upload/image", middleware.RequirePermission("content:file:upload"), fc.UploadImage)
	}
}
