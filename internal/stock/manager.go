package stock

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"stocker/internal/config"
	"stocker/internal/data"
	"stocker/internal/market"
	"stocker/pkg/models"
)

// Manager 股票管理器
type Manager struct {
	config    *config.Config
	provider  data.StockProvider
	cache     *data.Cache
	watchlist []models.Stock
	mutex     sync.RWMutex
	closeOnce sync.Once
}

// NewManager 创建新的股票管理器
func NewManager(cfg *config.Config) (*Manager, error) {
	primaryProvider, err := newProviderFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	provider := data.NewProviderChain(primaryProvider, nil)

	// 备用提供商可选（未注册则忽略）
	if cfg.API.FallbackProvider != "" && cfg.API.FallbackProvider != cfg.API.PrimaryProvider {
		if fallback, ferr := newProviderFromConfigWithName(cfg, cfg.API.FallbackProvider); ferr == nil {
			provider = data.NewProviderChain(primaryProvider, fallback)
		}
	}

	cache := data.NewCache(cfg.API.CacheDuration)

	manager := &Manager{
		config:    cfg,
		provider:  provider,
		cache:     cache,
		watchlist: cfg.Stocks.Watchlist,
	}

	// 从配置加载监控列表
	err = manager.loadFromConfig()
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	return manager, nil
}

// AddStock 添加股票到监控列表
func (m *Manager) AddStock(symbol string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	normalized, ok := market.NormalizeSymbol(symbol)
	if !ok {
		return fmt.Errorf("不支持的股票代码: %s", symbol)
	}

	// 验证股票代码
	valid, name := m.provider.ValidateStock(normalized)
	if !valid {
		return fmt.Errorf("无效的股票代码: %s", normalized)
	}

	// 检查是否已存在
	for _, stock := range m.watchlist {
		if stock.Symbol == normalized {
			return fmt.Errorf("股票 %s 已在监控列表中", normalized)
		}
	}

	// 添加到监控列表
	newStock := models.Stock{
		Symbol:  normalized,
		Name:    name,
		Market:  string(models.MarketCN),
		AddedAt: time.Now(),
	}

	m.watchlist = append(m.watchlist, newStock)

	// 保存配置
	return m.saveToConfig()
}

// RemoveStock 从监控列表删除股票
func (m *Manager) RemoveStock(symbol string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 查找股票
	index := -1
	for i, stock := range m.watchlist {
		if stock.Symbol == symbol {
			index = i
			break
		}
	}

	if index == -1 {
		return fmt.Errorf("股票 %s 不在监控列表中", symbol)
	}

	// 删除股票
	removedStock := m.watchlist[index]
	m.watchlist = append(m.watchlist[:index], m.watchlist[index+1:]...)

	// 清除缓存
	m.cache.Delete(symbol)

	// 保存配置
	err := m.saveToConfig()
	if err != nil {
		// 回滚删除操作
		m.watchlist = append(m.watchlist[:index], append([]models.Stock{removedStock}, m.watchlist[index:]...)...)
		return fmt.Errorf("保存配置失败: %w", err)
	}

	return nil
}

// RemoveStockByIndex 根据序号删除股票
func (m *Manager) RemoveStockByIndex(index int) error {
	m.mutex.RLock()
	if index < 0 || index >= len(m.watchlist) {
		m.mutex.RUnlock()
		return fmt.Errorf("无效的序号: %d", index+1) // 显示从1开始
	}
	symbol := m.watchlist[index].Symbol
	m.mutex.RUnlock()

	return m.RemoveStock(symbol)
}

// GetWatchlist 获取监控列表
func (m *Manager) GetWatchlist() []models.Stock {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 返回副本，避免外部修改
	watchlist := make([]models.Stock, len(m.watchlist))
	copy(watchlist, m.watchlist)
	return watchlist
}

// GetStockData 获取股票数据（带缓存）
func (m *Manager) GetStockData(symbol string) (*models.StockData, error) {
	// 先从缓存获取
	if cached, exists := m.cache.Get(symbol); exists {
		return cached, nil
	}

	// 从API获取
	stockData, err := m.provider.GetStockData(symbol)
	if err != nil {
		return nil, err
	}

	// 存入缓存
	m.cache.Set(symbol, stockData)
	return stockData, nil
}

// GetAllStockData 获取所有监控股票的数据（批量获取+按配置文件顺序排序）
func (m *Manager) GetAllStockData() ([]*models.StockData, error) {
	symbols := m.getWatchlistSymbols()
	if len(symbols) == 0 {
		return []*models.StockData{}, nil
	}

	// 批量获取数据
	allData, err := m.provider.BatchGetStockData(symbols)
	if err != nil {
		return nil, err
	}

	// 更新缓存
	for _, data := range allData {
		m.cache.Set(data.Symbol, data)
	}

	// 按配置文件顺序重新排序
	watchlist := m.GetWatchlist()
	symbolToIndex := make(map[string]int)
	for i, stock := range watchlist {
		symbolToIndex[stock.Symbol] = i
	}

	sort.Slice(allData, func(i, j int) bool {
		return symbolToIndex[allData[i].Symbol] < symbolToIndex[allData[j].Symbol]
	})

	return allData, nil
}

// SearchStock 搜索股票
func (m *Manager) SearchStock(keyword string) ([]models.SearchStock, error) {
	return m.provider.SearchStock(keyword)
}

// ReloadConfig 重新加载配置
func (m *Manager) ReloadConfig() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.loadFromConfig()
}

// SaveConfig 保存当前配置
func (m *Manager) SaveConfig() error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.saveToConfig()
}

// GetStats 获取统计信息
func (m *Manager) GetStats() ManagerStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return ManagerStats{
		TotalStocks: len(m.watchlist),
		CacheSize:   m.cache.Size(),
		LastUpdated: time.Now(),
	}
}

// loadFromConfig 从配置文件加载监控列表
func (m *Manager) loadFromConfig() error {
	if m.config == nil {
		return fmt.Errorf("配置为空")
	}

	var filtered []models.Stock
	for _, stock := range m.config.Stocks.Watchlist {
		norm, ok := market.NormalizeSymbol(stock.Symbol)
		if !ok {
			continue
		}
		stock.Symbol = norm
		stock.Market = string(models.MarketCN)
		filtered = append(filtered, stock)
	}

	m.watchlist = filtered
	return nil
}

func newProviderFromConfig(cfg *config.Config) (data.StockProvider, error) {
	name := cfg.API.PrimaryProvider
	if name == "" {
		name = "sina"
	}
	return newProviderFromConfigWithName(cfg, name)
}

func newProviderFromConfigWithName(cfg *config.Config, name string) (data.StockProvider, error) {
	factory, exists := data.GetProviderFactory(name)
	if !exists {
		return nil, fmt.Errorf("未注册的提供商: %s", name)
	}

	timeoutSec, retryDelaySec, cacheSec := normalizeGlobalAPIConfig(cfg)
	retryCount := cfg.API.RetryCount
	maxConcurrent := cfg.API.MaxConcurrent
	headers := map[string]string{}

	// 提供商专用配置（可选）
	if cfg.API.Providers != nil {
		if raw, ok := cfg.API.Providers[name]; ok {
			if pm, ok := toStringMap(raw); ok {
				if v, ok := getInt(pm, "timeout"); ok {
					timeoutSec = v
				}
				if v, ok := getInt(pm, "retry_delay"); ok {
					retryDelaySec = v
				}
				if v, ok := getInt(pm, "cache_duration"); ok {
					cacheSec = v
				}
				if v, ok := getInt(pm, "retry_count"); ok {
					retryCount = v
				}
				if v, ok := getInt(pm, "max_concurrent"); ok {
					maxConcurrent = v
				}
				if h, ok := getStringMap(pm, "headers"); ok {
					headers = h
				}
			}
		}
	}

	providerConfig := data.ProviderConfig{
		Name:          name,
		Timeout:       timeoutSec,
		RetryCount:    retryCount,
		RetryDelay:    retryDelaySec,
		MaxConcurrent: maxConcurrent,
		CacheDuration: cacheSec,
		Headers:       headers,
	}

	return factory.CreateProvider(providerConfig)
}

func normalizeGlobalAPIConfig(cfg *config.Config) (timeoutSec, retryDelaySec, cacheSec int) {
	timeoutSec = int(cfg.API.Timeout / time.Second)
	if timeoutSec <= 0 {
		timeoutSec = 5
	}
	retryDelaySec = int(cfg.API.RetryDelay / time.Second)
	if retryDelaySec <= 0 {
		retryDelaySec = 1
	}
	cacheSec = int(cfg.API.CacheDuration / time.Second)
	if cacheSec <= 0 {
		cacheSec = 2
	}
	return timeoutSec, retryDelaySec, cacheSec
}

func toStringMap(v interface{}) (map[string]interface{}, bool) {
	switch m := v.(type) {
	case map[string]interface{}:
		return m, true
	case map[interface{}]interface{}:
		out := make(map[string]interface{})
		for k, v := range m {
			ks, ok := k.(string)
			if !ok {
				continue
			}
			out[ks] = v
		}
		return out, true
	default:
		return nil, false
	}
}

func getInt(m map[string]interface{}, key string) (int, bool) {
	raw, ok := m[key]
	if !ok {
		return 0, false
	}
	switch v := raw.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case float32:
		return int(v), true
	default:
		return 0, false
	}
}

func getStringMap(m map[string]interface{}, key string) (map[string]string, bool) {
	raw, ok := m[key]
	if !ok {
		return nil, false
	}
	switch v := raw.(type) {
	case map[string]string:
		return v, true
	case map[string]interface{}:
		out := make(map[string]string)
		for k, iv := range v {
			if sv, ok := iv.(string); ok {
				out[k] = sv
			}
		}
		return out, true
	case map[interface{}]interface{}:
		out := make(map[string]string)
		for k, iv := range v {
			ks, ok := k.(string)
			if !ok {
				continue
			}
			if sv, ok := iv.(string); ok {
				out[ks] = sv
			}
		}
		return out, true
	default:
		return nil, false
	}
}

// saveToConfig 保存监控列表到配置文件
func (m *Manager) saveToConfig() error {
	if m.config == nil {
		return fmt.Errorf("配置为空")
	}

	// 更新配置
	m.config.Stocks.Watchlist = m.watchlist

	// 保存到文件
	return m.config.Save()
}

// getWatchlistSymbols 获取监控列表中的所有股票代码
func (m *Manager) getWatchlistSymbols() []string {
	symbols := make([]string, 0, len(m.watchlist))
	for _, stock := range m.watchlist {
		symbols = append(symbols, stock.Symbol)
	}
	return symbols
}

// determineMarket 根据股票代码判断市场类型
// GetSortedWatchlist 获取按添加时间排序的监控列表
func (m *Manager) GetSortedWatchlist() []models.Stock {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 创建副本并按添加时间排序
	watchlist := make([]models.Stock, len(m.watchlist))
	copy(watchlist, m.watchlist)

	sort.Slice(watchlist, func(i, j int) bool {
		return watchlist[i].AddedAt.Before(watchlist[j].AddedAt)
	})

	return watchlist
}

// ClearCache 清空缓存
func (m *Manager) ClearCache() {
	m.cache.Clear()
}

// GetCacheStats 获取缓存统计信息
func (m *Manager) GetCacheStats() data.CacheStats {
	return m.cache.GetCacheStats()
}

// RefreshStockData 刷新指定股票的数据
func (m *Manager) RefreshStockData(symbol string) (*models.StockData, error) {
	// 清除缓存
	m.cache.Delete(symbol)

	// 重新获取数据
	return m.GetStockData(symbol)
}

// Close 关闭管理器，清理资源
func (m *Manager) Close() {
	m.closeOnce.Do(func() {
		if m.cache != nil {
			m.cache.Stop()
		}
	})
}

// ManagerStats 管理器统计信息
type ManagerStats struct {
	TotalStocks int       `json:"total_stocks"`
	CacheSize   int       `json:"cache_size"`
	LastUpdated time.Time `json:"last_updated"`
}
