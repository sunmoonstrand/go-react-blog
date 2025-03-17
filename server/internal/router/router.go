package router

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	v1 "github.com/sunmoonstrand/go-react-blog/server/api/v1"
	"github.com/sunmoonstrand/go-react-blog/server/internal/config"
	"github.com/sunmoonstrand/go-react-blog/server/internal/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// InitRouter 初始化路由
func InitRouter(cfg *config.Config) *gin.Engine {
	// 设置运行模式
	gin.SetMode(cfg.Server.Mode)

	// 创建路由
	r := gin.New()

	// 使用中间件
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.RequestID())

	// 配置CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.Server.Cors.AllowedOrigins,
		AllowMethods:     cfg.Server.Cors.AllowedMethods,
		AllowHeaders:     cfg.Server.Cors.AllowedHeaders,
		ExposeHeaders:    cfg.Server.Cors.ExposedHeaders,
		AllowCredentials: cfg.Server.Cors.AllowCredentials,
		MaxAge:           time.Duration(cfg.Server.Cors.MaxAge) * time.Second,
	}))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// 静态文件服务
	r.Static("/uploads", "./uploads")
	r.StaticFile("/favicon.ico", "./static/favicon.ico")

	// Swagger文档
	if cfg.Swagger.Enabled {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// 初始化控制器
	authController := v1.NewAuthController()
	userController := v1.NewUserController()
	roleController := v1.NewRoleController()
	articleController := v1.NewArticleController()
	categoryController := v1.NewCategoryController()
	tagController := v1.NewTagController()
	commentController := v1.NewCommentController()
	configController := v1.NewConfigController()
	fileController := v1.NewFileController()

	// API路由组 - 前台接口
	apiV1 := r.Group("/api/v1")
	{
		// 无需认证的路由
		publicRoutes(apiV1, authController, articleController, categoryController, tagController, commentController)

		// 需要认证的路由
		authRoutes := apiV1.Group("")
		authRoutes.Use(middleware.JWTAuth(cfg.Server.JWTSecret))
		{
			// 用户相关路由
			userRoutes(authRoutes, userController)

			// 内容相关路由
			contentRoutes(authRoutes, articleController, categoryController, tagController, commentController)

			// 系统相关路由
			systemRoutes(authRoutes, fileController)
		}
	}

	// 后台管理API路由组
	adminV1 := r.Group("/admin/api/v1")
	{
		// 无需认证的后台路由
		adminPublicRoutes(adminV1, authController)

		// 需要认证的后台路由
		adminAuthRoutes := adminV1.Group("")
		adminAuthRoutes.Use(middleware.JWTAuth(cfg.Server.JWTSecret))
		adminAuthRoutes.Use(middleware.RBACAuth())
		{
			// 用户管理路由
			adminUserRoutes(adminAuthRoutes, userController, roleController)

			// 内容管理路由
			adminContentRoutes(adminAuthRoutes, articleController, categoryController, tagController, commentController, fileController)

			// 系统管理路由
			adminSystemRoutes(adminAuthRoutes, configController)
		}
	}

	// 404处理
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "API路由不存在",
		})
	})

	return r
}

// publicRoutes 注册公开路由
func publicRoutes(rg *gin.RouterGroup, authCtrl *v1.AuthController, articleCtrl *v1.ArticleController,
	categoryCtrl *v1.CategoryController, tagCtrl *v1.TagController, commentCtrl *v1.CommentController) {

	// 认证相关
	authGroup := rg.Group("/auth")
	{
		authCtrl.RegisterPublicRoutes(authGroup)
	}

	// 文章相关
	articleGroup := rg.Group("/article")
	{
		articleCtrl.RegisterPublicRoutes(articleGroup)
	}

	// 分类相关
	categoryGroup := rg.Group("/category")
	{
		categoryCtrl.RegisterPublicRoutes(categoryGroup)
	}

	// 标签相关
	tagGroup := rg.Group("/tag")
	{
		tagCtrl.RegisterPublicRoutes(tagGroup)
	}

	// 评论相关
	commentGroup := rg.Group("/comment")
	{
		commentCtrl.RegisterPublicRoutes(commentGroup)
	}
}

// userRoutes 注册用户相关路由
func userRoutes(rg *gin.RouterGroup, userCtrl *v1.UserController) {
	userGroup := rg.Group("/user")
	{
		userCtrl.RegisterRoutes(userGroup)
	}
}

// contentRoutes 注册内容相关路由
func contentRoutes(rg *gin.RouterGroup, articleCtrl *v1.ArticleController,
	categoryCtrl *v1.CategoryController, tagCtrl *v1.TagController, commentCtrl *v1.CommentController) {

	// 文章相关
	articleGroup := rg.Group("/article")
	{
		articleCtrl.RegisterRoutes(articleGroup)
	}

	// 分类相关
	categoryGroup := rg.Group("/category")
	{
		categoryCtrl.RegisterRoutes(categoryGroup)
	}

	// 标签相关
	tagGroup := rg.Group("/tag")
	{
		tagCtrl.RegisterRoutes(tagGroup)
	}

	// 评论相关
	commentGroup := rg.Group("/comment")
	{
		commentCtrl.RegisterRoutes(commentGroup)
	}
}

// systemRoutes 注册系统相关路由
func systemRoutes(rg *gin.RouterGroup, fileCtrl *v1.FileController) {
	// 文件上传
	fileGroup := rg.Group("/file")
	{
		fileCtrl.RegisterRoutes(fileGroup)
	}
}

// adminPublicRoutes 注册后台公开路由
func adminPublicRoutes(rg *gin.RouterGroup, authCtrl *v1.AuthController) {
	// 认证相关
	authGroup := rg.Group("/auth")
	{
		authCtrl.RegisterAdminPublicRoutes(authGroup)
	}
}

// adminUserRoutes 注册后台用户管理路由
func adminUserRoutes(rg *gin.RouterGroup, userCtrl *v1.UserController, roleCtrl *v1.RoleController) {
	// 用户管理
	userGroup := rg.Group("/user")
	{
		userCtrl.RegisterAdminRoutes(userGroup)
	}

	// 角色管理
	roleGroup := rg.Group("/role")
	{
		roleCtrl.RegisterRoutes(roleGroup)
	}
}

// adminContentRoutes 注册后台内容管理路由
func adminContentRoutes(rg *gin.RouterGroup, articleCtrl *v1.ArticleController,
	categoryCtrl *v1.CategoryController, tagCtrl *v1.TagController,
	commentCtrl *v1.CommentController, fileCtrl *v1.FileController) {

	// 文章管理
	articleGroup := rg.Group("/article")
	{
		articleCtrl.RegisterAdminRoutes(articleGroup)
	}

	// 分类管理
	categoryGroup := rg.Group("/category")
	{
		categoryCtrl.RegisterAdminRoutes(categoryGroup)
	}

	// 标签管理
	tagGroup := rg.Group("/tag")
	{
		tagCtrl.RegisterAdminRoutes(tagGroup)
	}

	// 评论管理
	commentGroup := rg.Group("/comment")
	{
		commentCtrl.RegisterAdminRoutes(commentGroup)
	}

	// 文件管理
	fileGroup := rg.Group("/file")
	{
		fileCtrl.RegisterAdminRoutes(fileGroup)
	}
}

// adminSystemRoutes 注册后台系统管理路由
func adminSystemRoutes(rg *gin.RouterGroup, configCtrl *v1.ConfigController) {
	// 系统配置
	configGroup := rg.Group("/config")
	{
		configCtrl.RegisterRoutes(configGroup)
	}
}
