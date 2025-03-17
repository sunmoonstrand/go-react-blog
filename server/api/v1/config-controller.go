package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/sunmoonstrand/go-react-blog/server/internal/logger"
	"github.com/sunmoonstrand/go-react-blog/server/internal/middleware"
	"github.com/sunmoonstrand/go-react-blog/server/internal/model"
	"github.com/sunmoonstrand/go-react-blog/server/internal/utils/resp"
)

// ConfigController 系统配置控制器
type ConfigController struct {
	configModel *model.ConfigModel
}

// NewConfigController 创建系统配置控制器实例
func NewConfigController(configModel *model.ConfigModel) *ConfigController {
	return &ConfigController{
		configModel: configModel,
	}
}

// UpdateSiteConfigRequest 站点配置请求结构
type UpdateSiteConfigRequest struct {
	SiteName        string `json:"site_name" binding:"required,min=2,max=50"`
	SiteDescription string `json:"site_description" binding:"max=200"`
	SiteKeywords    string `json:"site_keywords" binding:"max=200"`
	SiteLogo        string `json:"site_logo"`
	SiteFavicon     string `json:"site_favicon"`
	SiteNotice      string `json:"site_notice" binding:"max=500"`
	SiteFooter      string `json:"site_footer" binding:"max=1000"`
	SEOTitle        string `json:"seo_title" binding:"max=100"`
	SEODescription  string `json:"seo_description" binding:"max=200"`
	SEOKeywords     string `json:"seo_keywords" binding:"max=200"`
}

// UpdateCommentConfigRequest 评论配置请求结构
type UpdateCommentConfigRequest struct {
	CommentEnabled  bool `json:"comment_enabled"`
	CommentAudit    bool `json:"comment_audit"`
	CommentCaptcha  bool `json:"comment_captcha"`
	CommentInterval int  `json:"comment_interval" binding:"min=0,max=3600"`
}

// UpdateUploadConfigRequest 上传配置请求结构
type UpdateUploadConfigRequest struct {
	UploadMaxSize        int64    `json:"upload_max_size" binding:"min=1,max=100"`
	UploadAllowedFormats []string `json:"upload_allowed_formats" binding:"required,min=1"`
	UploadSavePath       string   `json:"upload_save_path" binding:"required"`
	UploadAccessUrl      string   `json:"upload_access_url" binding:"required"`
	UploadProvider       string   `json:"upload_provider" binding:"required,oneof=local oss cos qiniu"`
	UploadOssConfig      string   `json:"upload_oss_config"`
}

// UpdateEmailConfigRequest 邮件配置请求结构
type UpdateEmailConfigRequest struct {
	EmailEnabled  bool   `json:"email_enabled"`
	EmailHost     string `json:"email_host" binding:"required_if=EmailEnabled true"`
	EmailPort     int    `json:"email_port" binding:"required_if=EmailEnabled true,min=1,max=65535"`
	EmailUsername string `json:"email_username" binding:"required_if=EmailEnabled true"`
	EmailPassword string `json:"email_password" binding:"required_if=EmailEnabled true"`
	EmailFrom     string `json:"email_from" binding:"required_if=EmailEnabled true,email"`
}

// GetSystemConfig 获取系统配置
// @Summary 获取系统配置
// @Description 获取系统所有配置信息
// @Tags 系统配置
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} resp.Response "返回系统配置信息"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/config [get]
func (cc *ConfigController) GetSystemConfig(c *gin.Context) {
	// 获取站点配置
	siteConfig, err := cc.configModel.GetSiteConfig()
	if err != nil {
		logger.Error("获取站点配置失败", "error", err)
		resp.FailWithMsg(c, "获取系统配置失败")
		return
	}

	// 获取评论配置
	commentConfig, err := cc.configModel.GetCommentConfig()
	if err != nil {
		logger.Error("获取评论配置失败", "error", err)
		resp.FailWithMsg(c, "获取系统配置失败")
		return
	}

	// 获取上传配置
	uploadConfig, err := cc.configModel.GetUploadConfig()
	if err != nil {
		logger.Error("获取上传配置失败", "error", err)
		resp.FailWithMsg(c, "获取系统配置失败")
		return
	}

	// 获取邮件配置
	emailConfig, err := cc.configModel.GetEmailConfig()
	if err != nil {
		logger.Error("获取邮件配置失败", "error", err)
		resp.FailWithMsg(c, "获取系统配置失败")
		return
	}

	// 构建响应数据
	resp.OkWithData(c, gin.H{
		"site":    siteConfig,
		"comment": commentConfig,
		"upload":  uploadConfig,
		"email":   emailConfig,
	})
}

// GetPublicConfig 获取公共配置
// @Summary 获取公共配置
// @Description 获取公开的系统配置信息，无需登录权限
// @Tags 系统配置
// @Accept json
// @Produce json
// @Success 200 {object} resp.Response "返回公共配置信息"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/config/public [get]
func (cc *ConfigController) GetPublicConfig(c *gin.Context) {
	// 获取站点公共配置
	siteConfig, err := cc.configModel.GetPublicSiteConfig()
	if err != nil {
		logger.Error("获取公共站点配置失败", "error", err)
		resp.FailWithMsg(c, "获取系统配置失败")
		return
	}

	// 获取评论公共配置
	commentConfig, err := cc.configModel.GetPublicCommentConfig()
	if err != nil {
		logger.Error("获取公共评论配置失败", "error", err)
		resp.FailWithMsg(c, "获取系统配置失败")
		return
	}

	// 构建响应数据
	resp.OkWithData(c, gin.H{
		"site":    siteConfig,
		"comment": commentConfig,
	})
}

// UpdateSiteConfig 更新站点配置
// @Summary 更新站点配置
// @Description 更新站点相关配置信息
// @Tags 系统配置
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body UpdateSiteConfigRequest true "站点配置信息"
// @Success 200 {object} resp.Response "更新成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/config/site [put]
func (cc *ConfigController) UpdateSiteConfig(c *gin.Context) {
	var req UpdateSiteConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.FailWithValidation(c, err)
		return
	}

	// 构建配置数据
	config := map[string]string{
		"site_name":        req.SiteName,
		"site_description": req.SiteDescription,
		"site_keywords":    req.SiteKeywords,
		"site_logo":        req.SiteLogo,
		"site_favicon":     req.SiteFavicon,
		"site_notice":      req.SiteNotice,
		"site_footer":      req.SiteFooter,
		"seo_title":        req.SEOTitle,
		"seo_description":  req.SEODescription,
		"seo_keywords":     req.SEOKeywords,
	}

	// 更新配置
	if err := cc.configModel.UpdateSiteConfig(config); err != nil {
		logger.Error("更新站点配置失败", "error", err)
		resp.FailWithMsg(c, "更新站点配置失败")
		return
	}

	resp.OkWithMsg(c, "更新站点配置成功")
}

// UpdateCommentConfig 更新评论配置
// @Summary 更新评论配置
// @Description 更新评论相关配置信息
// @Tags 系统配置
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body UpdateCommentConfigRequest true "评论配置信息"
// @Success 200 {object} resp.Response "更新成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/config/comment [put]
func (cc *ConfigController) UpdateCommentConfig(c *gin.Context) {
	var req UpdateCommentConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.FailWithValidation(c, err)
		return
	}

	// 构建配置数据
	config := make(map[string]interface{})
	config["comment_enabled"] = req.CommentEnabled
	config["comment_audit"] = req.CommentAudit
	config["comment_captcha"] = req.CommentCaptcha
	config["comment_interval"] = req.CommentInterval

	// 更新配置
	if err := cc.configModel.UpdateCommentConfig(config); err != nil {
		logger.Error("更新评论配置失败", "error", err)
		resp.FailWithMsg(c, "更新评论配置失败")
		return
	}

	resp.OkWithMsg(c, "更新评论配置成功")
}

// UpdateUploadConfig 更新上传配置
// @Summary 更新上传配置
// @Description 更新文件上传相关配置信息
// @Tags 系统配置
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body UpdateUploadConfigRequest true "上传配置信息"
// @Success 200 {object} resp.Response "更新成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/config/upload [put]
func (cc *ConfigController) UpdateUploadConfig(c *gin.Context) {
	var req UpdateUploadConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.FailWithValidation(c, err)
		return
	}

	// 构建配置数据
	config := make(map[string]interface{})
	config["upload_max_size"] = req.UploadMaxSize
	config["upload_allowed_formats"] = req.UploadAllowedFormats
	config["upload_save_path"] = req.UploadSavePath
	config["upload_access_url"] = req.UploadAccessUrl
	config["upload_provider"] = req.UploadProvider
	config["upload_oss_config"] = req.UploadOssConfig

	// 更新配置
	if err := cc.configModel.UpdateUploadConfig(config); err != nil {
		logger.Error("更新上传配置失败", "error", err)
		resp.FailWithMsg(c, "更新上传配置失败")
		return
	}

	resp.OkWithMsg(c, "更新上传配置成功")
}

// UpdateEmailConfig 更新邮件配置
// @Summary 更新邮件配置
// @Description 更新邮件发送相关配置信息
// @Tags 系统配置
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body UpdateEmailConfigRequest true "邮件配置信息"
// @Success 200 {object} resp.Response "更新成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/config/email [put]
func (cc *ConfigController) UpdateEmailConfig(c *gin.Context) {
	var req UpdateEmailConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.FailWithValidation(c, err)
		return
	}

	// 如果启用邮件，验证邮件配置
	if req.EmailEnabled {
		// 测试邮件连接
		err := cc.configModel.TestEmailConnection(req.EmailHost, req.EmailPort, req.EmailUsername, req.EmailPassword)
		if err != nil {
			logger.Error("邮件服务器连接测试失败", "error", err)
			resp.FailWithMsg(c, "邮件服务器连接失败，请检查配置: "+err.Error())
			return
		}
	}

	// 构建配置数据
	config := make(map[string]interface{})
	config["email_enabled"] = req.EmailEnabled
	config["email_host"] = req.EmailHost
	config["email_port"] = req.EmailPort
	config["email_username"] = req.EmailUsername
	config["email_password"] = req.EmailPassword
	config["email_from"] = req.EmailFrom

	// 更新配置
	if err := cc.configModel.UpdateEmailConfig(config); err != nil {
		logger.Error("更新邮件配置失败", "error", err)
		resp.FailWithMsg(c, "更新邮件配置失败")
		return
	}

	resp.OkWithMsg(c, "更新邮件配置成功")
}

// TestEmailSend 测试邮件发送
// @Summary 测试邮件发送
// @Description 测试邮件配置是否能成功发送邮件
// @Tags 系统配置
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param to query string true "收件人邮箱"
// @Success 200 {object} resp.Response "发送成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/config/email/test [post]
func (cc *ConfigController) TestEmailSend(c *gin.Context) {
	to := c.Query("to")
	if to == "" {
		resp.FailWithMsg(c, "收件人邮箱不能为空")
		return
	}

	// 获取邮件配置
	emailConfig, err := cc.configModel.GetEmailConfig()
	if err != nil {
		logger.Error("获取邮件配置失败", "error", err)
		resp.FailWithMsg(c, "测试邮件发送失败")
		return
	}

	// 检查邮件是否启用
	if !emailConfig.Enabled {
		resp.FailWithMsg(c, "邮件服务未启用，请先启用邮件服务")
		return
	}

	// 发送测试邮件
	err = cc.configModel.SendTestEmail(to)
	if err != nil {
		logger.Error("发送测试邮件失败", "to", to, "error", err)
		resp.FailWithMsg(c, "发送测试邮件失败: "+err.Error())
		return
	}

	resp.OkWithMsg(c, "测试邮件发送成功")
}

// ClearCache 清除系统缓存
// @Summary 清除系统缓存
// @Description 清除系统各类缓存
// @Tags 系统配置
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param cache_type query string false "缓存类型" Enums(all, config, article, category, tag, user)
// @Success 200 {object} resp.Response "清除成功"
// @Failure 401 {object} resp.Response "未授权"
// @Failure 403 {object} resp.Response "无权限"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /api/v1/config/cache/clear [post]
func (cc *ConfigController) ClearCache(c *gin.Context) {
	cacheType := c.DefaultQuery("cache_type", "all")

	// 清除缓存
	err := cc.configModel.ClearCache(cacheType)
	if err != nil {
		logger.Error("清除缓存失败", "cache_type", cacheType, "error", err)
		resp.FailWithMsg(c, "清除缓存失败")
		return
	}

	resp.OkWithMsg(c, "清除缓存成功")
}

// RegisterRoutes 注册路由
func (cc *ConfigController) RegisterRoutes(router *gin.RouterGroup) {
	configGroup := router.Group("/config")
	{
		// 获取系统配置(需要权限)
		configGroup.GET("", middleware.RequirePermission("system:config:info"), cc.GetSystemConfig)

		// 获取公共配置(无需权限)
		configGroup.GET("/public", cc.GetPublicConfig)

		// 更新站点配置
		configGroup.PUT("/site", middleware.RequirePermission("system:config:edit"), cc.UpdateSiteConfig)

		// 更新评论配置
		configGroup.PUT("/comment", middleware.RequirePermission("system:config:edit"), cc.UpdateCommentConfig)

		// 更新上传配置
		configGroup.PUT("/upload", middleware.RequirePermission("system:config:edit"), cc.UpdateUploadConfig)

		// 更新邮件配置
		configGroup.PUT("/email", middleware.RequirePermission("system:config:edit"), cc.UpdateEmailConfig)

		// 测试邮件发送
		configGroup.POST("/email/test", middleware.RequirePermission("system:config:edit"), cc.TestEmailSend)

		// 清除系统缓存
		configGroup.POST("/cache/clear", middleware.RequirePermission("system:config:edit"), cc.ClearCache)
	}
}
