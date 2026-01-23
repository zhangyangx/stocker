package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
	"stocker/pkg/models"
)

// Config 表示应用程序的配置结构
type Config struct {
	App         App         `yaml:"app"`
	Preferences Preferences `yaml:"preferences"`
	API         API         `yaml:"api"`
	Stocks      Stocks      `yaml:"stocks"`
}

// App 应用程序配置
type App struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

// Preferences 用户偏好设置
type Preferences struct {
	RefreshInterval int  `yaml:"refresh_interval"`
	SimpleMode      bool `yaml:"simple_mode"`
}

// API API配置
type API struct {
	Timeout          time.Duration          `yaml:"timeout"`
	RetryCount       int                    `yaml:"retry_count"`
	RetryDelay       time.Duration          `yaml:"retry_delay"`
	MaxConcurrent    int                    `yaml:"max_concurrent"`
	CacheDuration    time.Duration          `yaml:"cache_duration"`
	PrimaryProvider  string                 `yaml:"primary_provider"`
	FallbackProvider string                 `yaml:"fallback_provider"`
	Providers        map[string]interface{} `yaml:"providers"` // 各提供商的特定配置
}

// Stocks 股票配置
type Stocks struct {
	Watchlist []models.Stock `yaml:"watchlist"`
}

// Stock 单个股票配置（作为存储结构）
type Stock struct {
	Symbol  string    `yaml:"symbol"`
	Name    string    `yaml:"name"`
	Market  string    `yaml:"market"`
	AddedAt time.Time `yaml:"added_at"`
}

const (
	configFileName = "config.yaml"
	appName        = "stocker"
)

// GetConfigDir 获取配置文件目录
func GetConfigDir() (string, error) {
	if dir := os.Getenv("STOCKER_CONFIG_DIR"); dir != "" {
		return dir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(home, ".config", appName)
	return configDir, nil
}

// GetConfigPath 获取配置文件的完整路径
func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, configFileName), nil
}

// Load 加载配置文件
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 配置文件不存在，创建默认配置
		if err := createDefaultConfig(configPath); err != nil {
			return nil, err
		}
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 设置默认值
	setDefaults(&config)

	return &config, nil
}

// Save 保存配置到文件
func (c *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// 确保配置目录存在
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 序列化配置
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	// 写入配置文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// createDefaultConfig 创建默认配置文件
func createDefaultConfig(configPath string) error {
	config := GetDefaultConfig()
	return config.Save()
}

// setDefaults 设置默认配置值
func setDefaults(c *Config) {
	defaultConfig := GetDefaultConfig()

	if c.App.Name == "" {
		c.App.Name = defaultConfig.App.Name
	}
	if c.App.Version == "" {
		c.App.Version = defaultConfig.App.Version
	}
	if c.Preferences.RefreshInterval <= 0 {
		c.Preferences.RefreshInterval = defaultConfig.Preferences.RefreshInterval
	}
	// SimpleMode 默认为 false，不需要特别处理
	// 如果配置文件中没有该字段，Go 的零值就是 false
	if c.API.Timeout == 0 {
		c.API.Timeout = defaultConfig.API.Timeout
	}
	if c.API.RetryCount == 0 {
		c.API.RetryCount = defaultConfig.API.RetryCount
	}
	if c.API.RetryDelay == 0 {
		c.API.RetryDelay = defaultConfig.API.RetryDelay
	}
	if c.API.MaxConcurrent == 0 {
		c.API.MaxConcurrent = defaultConfig.API.MaxConcurrent
	}
	if c.API.CacheDuration == 0 {
		c.API.CacheDuration = defaultConfig.API.CacheDuration
	}
	if c.API.PrimaryProvider == "" {
		c.API.PrimaryProvider = defaultConfig.API.PrimaryProvider
	}
	if c.API.FallbackProvider == "" {
		c.API.FallbackProvider = defaultConfig.API.FallbackProvider
	}
	if c.API.Providers == nil {
		c.API.Providers = defaultConfig.API.Providers
	}
	if c.Stocks.Watchlist == nil {
		c.Stocks.Watchlist = defaultConfig.Stocks.Watchlist
	}
}
