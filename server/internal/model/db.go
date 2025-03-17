package model

import (
	"fmt"
	"time"

	"github.com/sunmoonstrand/go-react-blog/server/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// DB 全局数据库连接
var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB(config config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		config.Host,
		config.Port,
		config.Username,
		config.Password,
		config.DBName,
		config.SSLMode,
		config.Timezone,
	)

	// 配置GORM
	gormConfig := &gorm.Config{
		// 禁用默认事务
		SkipDefaultTransaction: true,
		// 命名策略
		NamingStrategy: schema.NamingStrategy{
			// 使用单数表名
			SingularTable: true,
		},
		// 日志配置
		Logger: logger.Default.LogMode(logger.Info),
		// 禁用外键约束
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, err
	}

	// 获取底层SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Second)

	// 设置全局DB
	DB = db

	return db, nil
}
