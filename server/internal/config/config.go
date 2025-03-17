package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config 总配置结构体
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Log      LogConfig      `mapstructure:"log"`
	Upload   UploadConfig   `mapstructure:"upload"`
	Swagger  SwaggerConfig  `mapstructure:"swagger"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host             string     `mapstructure:"host"`
	Port             int        `mapstructure:"port"`
	Mode             string     `mapstructure:"mode"`
	JWTSecret        string     `mapstructure:"jwt_secret"`
	JWTExpire        int        `mapstructure:"jwt_expire"`
	JWTIssuer        string     `mapstructure:"jwt_issuer"`
	JWTRefreshExpire int        `mapstructure:"jwt_refresh_expire"`
	Cors             CorsConfig `mapstructure:"cors"`
}

// CorsConfig CORS配置
type CorsConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	ExposedHeaders   []string `mapstructure:"exposed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	SSLMode         string `mapstructure:"sslmode"`
	Timezone        string `mapstructure:"timezone"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
	DialTimeout  int    `mapstructure:"dial_timeout"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level            string   `mapstructure:"level"`
	Format           string   `mapstructure:"format"`
	OutputPaths      []string `mapstructure:"output_paths"`
	ErrorOutputPaths []string `mapstructure:"error_output_paths"`
	MaxSize          int      `mapstructure:"max_size"`
	MaxAge           int      `mapstructure:"max_age"`
	MaxBackups       int      `mapstructure:"max_backups"`
	Compress         bool     `mapstructure:"compress"`
	LocalTime        bool     `mapstructure:"local_time"`
}

// UploadConfig 文件上传配置
type UploadConfig struct {
	MaxSize     int      `mapstructure:"max_size"`
	AllowTypes  []string `mapstructure:"allow_types"`
	SavePath    string   `mapstructure:"save_path"`
	URLPrefix   string   `mapstructure:"url_prefix"`
	StorageType string   `mapstructure:"storage_type"`
}

// SwaggerConfig Swagger文档配置
type SwaggerConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Title       string `mapstructure:"title"`
	Description string `mapstructure:"description"`
	Version     string `mapstructure:"version"`
	Host        string `mapstructure:"host"`
	BasePath    string `mapstructure:"base_path"`
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	viper.AddConfigPath(configPath)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// 环境变量替换
	viper.AutomaticEnv()
	viper.SetEnvPrefix("BLOG")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	// 将配置解析为结构体
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
