package data

import (
	"stocker/pkg/models"
	"time"
)

// StockProvider 股票数据提供商接口
type StockProvider interface {
	// GetStockData 获取单个股票数据
	GetStockData(symbol string) (*models.StockData, error)

	// BatchGetStockData 批量获取股票数据
	BatchGetStockData(symbols []string) ([]*models.StockData, error)

	// SearchStock 搜索股票
	SearchStock(keyword string) ([]models.SearchStock, error)

	// ValidateStock 验证股票代码是否有效
	ValidateStock(symbol string) (bool, string)

	// GetProviderName 获取提供商名称
	GetProviderName() string
}

// APIConfig API配置结构
type APIConfig struct {
	Timeout       time.Duration
	RetryCount    int
	RetryDelay    time.Duration
	MaxConcurrent int
	CacheDuration time.Duration
	Headers       map[string]string
}

// ProviderConfig 提供商配置结构
type ProviderConfig struct {
	Name          string
	Timeout       int
	RetryCount    int
	RetryDelay    int
	MaxConcurrent int
	CacheDuration int
	Headers       map[string]string
}

// ProviderFactory 提供商工厂接口
type ProviderFactory interface {
	// CreateProvider 创建提供商实例
	CreateProvider(config ProviderConfig) (StockProvider, error)

	// GetProviderName 获取提供商名称
	GetProviderName() string

	// IsSupported 检查是否支持某种股票
	IsSupported(symbol string) bool
}

// ProviderRegistry 提供商注册表
var ProviderRegistry = make(map[string]ProviderFactory)

// RegisterProvider 注册提供商
func RegisterProvider(name string, factory ProviderFactory) {
	if _, exists := ProviderRegistry[name]; exists {
		panic("provider already registered: " + name)
	}
	ProviderRegistry[name] = factory
}

// GetProviderFactory 获取提供商工厂
func GetProviderFactory(name string) (ProviderFactory, bool) {
	factory, exists := ProviderRegistry[name]
	return factory, exists
}

// GetProviders 获取所有已注册的提供商名称
func GetProviders() []string {
	var providers []string
	for name := range ProviderRegistry {
		providers = append(providers, name)
	}
	return providers
}
