package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sunmoonstrand/go-react-blog/server/internal/config"
	"github.com/sunmoonstrand/go-react-blog/server/internal/logger"
	"github.com/sunmoonstrand/go-react-blog/server/internal/model"
	"github.com/sunmoonstrand/go-react-blog/server/internal/router"

	"go.uber.org/zap"
)

// @title 博客系统API
// @version 1.0
// @description 博客系统后端API文档
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http https
func main() {
	// 初始化配置
	cfg, err := config.LoadConfig("./config")
	if err != nil {
		panic(fmt.Sprintf("加载配置文件失败: %v", err))
	}

	// 初始化日志
	log, err := logger.NewLogger(cfg.Log)
	if err != nil {
		panic(fmt.Sprintf("初始化日志失败: %v", err))
	}
	defer log.Sync()

	// 替换全局logger
	zap.ReplaceGlobals(log)

	// 初始化数据库连接
	db, err := model.InitDB(cfg.Database)
	if err != nil {
		log.Fatal("数据库连接失败", zap.Error(err))
	}

	// 获取底层sqlDB以便关闭
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("获取SQL DB实例失败", zap.Error(err))
	}
	defer sqlDB.Close()

	// 初始化Redis连接
	rdb, err := model.InitRedis(cfg.Redis)
	if err != nil {
		log.Fatal("Redis连接失败", zap.Error(err))
	}
	defer rdb.Close()

	// 初始化路由
	r := router.InitRouter(cfg)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: r,
	}

	// 启动HTTP服务器
	go func() {
		log.Info("启动服务器",
			zap.String("addr", server.Addr),
			zap.String("mode", cfg.Server.Mode),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("服务器启动失败", zap.Error(err))
		}
	}()

	// 等待中断信号优雅关闭服务器
	quit := make(chan os.Signal, 1)
	// kill (无参数) 默认发送 syscall.SIGTERM
	// kill -2 是 syscall.SIGINT
	// kill -9 是 syscall.SIGKILL 但无法被捕获，所以不需要添加它
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("正在关闭服务器...")

	// 设置关闭超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("服务器关闭异常", zap.Error(err))
	}

	log.Info("服务器已关闭")
}
