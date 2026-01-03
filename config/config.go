package config

import (
	"log"
	"os"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Upload   UploadConfig
	Session  SessionConfig
	Nav      NavConfig
}

type ServerConfig struct {
	Port string
	Mode string
}

type DatabaseConfig struct {
	Path string
}

type UploadConfig struct {
	Path         string
	MaxSize      int64
	AllowedTypes []string
}

type SessionConfig struct {
	Secret string
	MaxAge int
}

type NavConfig struct {
	JSONPath string // nav.json输出路径
}

var AppConfig *Config

func Init() {
	AppConfig = &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("SERVER_MODE", "release"),
		},
		Database: DatabaseConfig{
			Path: getEnv("DB_PATH", "./data/admin.db"),
		},
		Upload: UploadConfig{
			Path:         getEnv("UPLOAD_PATH", "./uploads"),
			MaxSize:      5 * 1024 * 1024, // 5MB
			AllowedTypes: []string{".png", ".jpg", ".jpeg", ".svg", ".gif", ".zip", ".rar", ".7z", ".pdf", ".doc", ".docx", ".xls", ".xlsx"},
		},
		Session: SessionConfig{
			Secret: getEnv("SESSION_SECRET", "nav-admin-secret-key-change-in-production"),
			MaxAge: 86400, // 24小时
		},
		Nav: NavConfig{
			JSONPath: getEnv("NAV_JSON_PATH", "./static/nav.json"),
		},
	}

	// 确保必要的目录存在
	os.MkdirAll(AppConfig.Upload.Path, 0755)
	os.MkdirAll("./data", 0755)

	log.Println("配置初始化完成")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
