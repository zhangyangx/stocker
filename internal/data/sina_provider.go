package data

import (
	"time"
)

// SinaProviderFactory 新浪提供商工厂
type SinaProviderFactory struct{}

// CreateProvider 创建新浪提供商实例
func (f *SinaProviderFactory) CreateProvider(config ProviderConfig) (StockProvider, error) {
	apiConfig := &APIConfig{
		Timeout:       time.Duration(config.Timeout) * time.Second,
		RetryCount:    config.RetryCount,
		RetryDelay:    time.Duration(config.RetryDelay) * time.Second,
		MaxConcurrent: config.MaxConcurrent,
		CacheDuration: time.Duration(config.CacheDuration) * time.Second,
	}

	return NewSinaClient(apiConfig), nil
}

// GetProviderName 获取提供商名称
func (f *SinaProviderFactory) GetProviderName() string {
	return "sina"
}

// IsSupported 检查是否支持某种股票
func (f *SinaProviderFactory) IsSupported(symbol string) bool {
	// 仅支持A股/ETF/指数
	if len(symbol) >= 2 {
		prefix := symbol[:2]
		return prefix == "sh" || prefix == "sz"
	}
	return false
}

// init 注册新浪提供商
func init() {
	RegisterProvider("sina", &SinaProviderFactory{})
}
