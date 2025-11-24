package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	WeChat   WeChatConfig   `mapstructure:"wechat"`
	Upload   UploadConfig   `mapstructure:"upload"`
}

type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	Mode         string `mapstructure:"mode"` // debug, release, test
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
	Type     string `mapstructure:"type"` // sqlite, mysql
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	DSN      string `mapstructure:"dsn"` // SQLite文件路径或MySQL DSN
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	Expiration int    `mapstructure:"expiration"` // 小时
}

type WeChatConfig struct {
	AppID     string `mapstructure:"app_id"`
	AppSecret string `mapstructure:"app_secret"`
	// AccountType: "official_account" (公众号) 或 "open_platform" (开放平台网站应用)
	// 默认: "open_platform"
	AccountType string `mapstructure:"account_type"`
	// Scope: "snsapi_base" (静默授权) 或 "snsapi_userinfo" (需要用户确认)
	// 仅当 AccountType 为 "official_account" 时有效
	// 默认: "snsapi_userinfo"
	Scope string `mapstructure:"scope"`
	// CallbackDomain: 回调域名（如：https://yourdomain.com），用于生成 redirect_uri
	// 如果设置了此值，将优先使用此域名，确保与微信后台配置的授权回调域名一致
	// 如果不设置，则从 Referer 头或查询参数获取
	CallbackDomain string `mapstructure:"callback_domain"`
}

type UploadConfig struct {
	StoragePath string   `mapstructure:"storage_path"` // 文件存储路径（相对路径或绝对路径）
	MaxFileSize int64    `mapstructure:"max_file_size"` // 最大文件大小（字节），默认 100MB
	AllowedTypes []string `mapstructure:"allowed_types"` // 允许的文件类型（MIME类型），空数组表示允许所有类型
}

var AppConfig *Config

func LoadConfig(configPath string) error {
	if configPath == "" {
		configPath = "config.yaml"
	}

	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 设置默认值
	setDefaults()

	// 读取环境变量
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Could not read config file: %v. Using defaults.", err)
	}

	AppConfig = &Config{}
	if err := viper.Unmarshal(AppConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

func setDefaults() {
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)

	viper.SetDefault("database.type", "sqlite")
	viper.SetDefault("database.dsn", "data.db")

	viper.SetDefault("jwt.secret", "your-secret-key-change-in-production")
	viper.SetDefault("jwt.expiration", 24)

	viper.SetDefault("wechat.app_id", "")
	viper.SetDefault("wechat.app_secret", "")
	viper.SetDefault("wechat.account_type", "open_platform") // open_platform 或 official_account
	viper.SetDefault("wechat.scope", "snsapi_userinfo")       // snsapi_base 或 snsapi_userinfo
	viper.SetDefault("wechat.callback_domain", "")            // 回调域名，如：https://yourdomain.com

	// 文件上传配置
	viper.SetDefault("upload.storage_path", "uploads")                    // 默认存储路径
	viper.SetDefault("upload.max_file_size", 100*1024*1024)               // 默认 100MB (104857600 字节)
	viper.SetDefault("upload.allowed_types", []string{})                  // 空数组表示允许所有类型
}

