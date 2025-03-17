package logger

import (
	"os"
	"time"

	"github.com/sunmoonstrand/go-react-blog/server/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// NewLogger 创建日志记录器
func NewLogger(config config.LogConfig) (*zap.Logger, error) {
	// 设置日志级别
	level := getLogLevel(config.Level)

	// 创建核心日志记录器
	cores := []zapcore.Core{}

	// 标准输出
	if len(config.OutputPaths) > 0 {
		for _, path := range config.OutputPaths {
			if path == "stdout" {
				cores = append(cores, createConsoleCore(level, config.Format))
			} else {
				cores = append(cores, createFileCore(path, level, config.Format, &config))
			}
		}
	} else {
		// 默认输出到控制台
		cores = append(cores, createConsoleCore(level, config.Format))
	}

	// 错误输出
	if len(config.ErrorOutputPaths) > 0 {
		for _, path := range config.ErrorOutputPaths {
			if path == "stderr" {
				// 只记录错误级别以上的日志
				cores = append(cores, createConsoleCore(zap.ErrorLevel, config.Format))
			} else {
				cores = append(cores, createFileCore(path, zap.ErrorLevel, config.Format, &config))
			}
		}
	}

	// 合并所有核心
	core := zapcore.NewTee(cores...)

	// 创建logger
	logger := zap.New(
		core,
		zap.AddCaller(),                   // 添加调用者信息
		zap.AddCallerSkip(1),              // 跳过日志封装函数
		zap.AddStacktrace(zap.ErrorLevel), // 为错误及以上级别添加堆栈跟踪
	)

	return logger, nil
}

// getLogLevel 将字符串日志级别转换为zap日志级别
func getLogLevel(levelStr string) zapcore.Level {
	switch levelStr {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "dpanic":
		return zap.DPanicLevel
	case "panic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}

// createConsoleCore 创建控制台日志核心
func createConsoleCore(level zapcore.Level, format string) zapcore.Core {
	// 判断编码器类型
	var encoder zapcore.Encoder
	if format == "json" {
		encoder = zapcore.NewJSONEncoder(getEncoderConfig())
	} else {
		encoder = zapcore.NewConsoleEncoder(getEncoderConfig())
	}

	// 创建输出
	syncer := zapcore.AddSync(os.Stdout)

	// 创建核心
	return zapcore.NewCore(encoder, syncer, level)
}

// createFileCore 创建文件日志核心
func createFileCore(path string, level zapcore.Level, format string, config *config.LogConfig) zapcore.Core {
	// 创建日志旋转配置
	lumberjackLogger := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
		LocalTime:  config.LocalTime,
	}

	// 判断编码器类型
	var encoder zapcore.Encoder
	if format == "json" {
		encoder = zapcore.NewJSONEncoder(getEncoderConfig())
	} else {
		encoder = zapcore.NewConsoleEncoder(getEncoderConfig())
	}

	// 创建输出
	syncer := zapcore.AddSync(lumberjackLogger)

	// 创建核心
	return zapcore.NewCore(encoder, syncer, level)
}

// getEncoderConfig 获取编码器配置
func getEncoderConfig() zapcore.EncoderConfig {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	encoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	return encoderConfig
}
