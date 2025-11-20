package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using default or environment variables")
	}
}

type OpenWrtConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Protocol string // http or https
}

// GetOpenWrtConfig 获取 OpenWrt 配置
func GetOpenWrtConfig() *OpenWrtConfig {
	protocol := getEnv("OPENWRT_PROTOCOL", "http")
	host := getEnv("OPENWRT_HOST", "192.168.1.1")
	port := getEnv("OPENWRT_PORT", "80")
	username := getEnv("OPENWRT_USERNAME", "root")
	password := getEnv("OPENWRT_PASSWORD", "")

	return &OpenWrtConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Protocol: protocol,
	}
}

// GetBaseURL 获取 OpenWrt 基础 URL
func (c *OpenWrtConfig) GetBaseURL() string {
	return c.Protocol + "://" + c.Host + ":" + c.Port
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
