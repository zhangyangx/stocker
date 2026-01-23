package config

import (
	"stocker/pkg/models"
	"time"
)

// 默认配置常量
const (
	DefaultRefreshInterval = 3               // 默认刷新间隔（秒）
	DefaultCacheDuration   = 2 * time.Second // 默认缓存时间（2秒）
	DefaultTimeout         = 5 * time.Second // 默认请求超时时间
	DefaultRetryCount      = 3               // 默认重试次数
	DefaultRetryDelay      = 1 * time.Second // 默认重试延迟
	MaxConcurrentRequests  = 10              // 最大并发请求数
)

// API相关配置常量
const (
	// 新浪财经API配置
	SinaAPIURL    = "https://hq.sinajs.cn/list="
	SinaSearchURL = "https://suggest3.sinajs.cn/suggest/key="
	SinaReferer   = "https://finance.sina.com.cn"
	SinaUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"

	// 腾讯财经API配置
	TencentAPIURL    = "https://qt.gtimg.cn/q="
	TencentSearchURL = "https://smartbox.gtimg.cn/s3/?v=2&t=all&c=1&q="
)

// 市场前缀配置
const (
	PrefixShanghai = "sh" // 上海A股前缀
	PrefixShenzhen = "sz" // 深圳A股前缀
)

// 默认配置结构
func GetDefaultConfig() *Config {
	return &Config{
		App: App{
			Name:    "Stocker",
			Version: "1.0.0-beta",
		},
		Preferences: Preferences{
			RefreshInterval: DefaultRefreshInterval,
		},
		API: API{
			Timeout:          DefaultTimeout,
			RetryCount:       DefaultRetryCount,
			RetryDelay:       DefaultRetryDelay,
			MaxConcurrent:    MaxConcurrentRequests,
			CacheDuration:    DefaultCacheDuration,
			PrimaryProvider:  "sina",
			FallbackProvider: "",
			Providers:        make(map[string]interface{}), // 初始化提供商配置映射
		},
		Stocks: Stocks{
			Watchlist: []models.Stock{},
		},
	}
}

// 获取默认的监控列表
func GetDefaultWatchlist() []Stock {
	return []Stock{}
}
