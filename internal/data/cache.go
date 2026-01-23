package data

import (
	"fmt"
	"stocker/pkg/models"
	"sync"
	"time"
)

// CacheEntry 缓存条目结构
type CacheEntry struct {
	Data      *models.StockData
	ExpiresAt time.Time
}

// Cache 缓存管理器
type Cache struct {
	entries       sync.Map // 使用 sync.Map 实现并发安全的缓存
	cacheDuration time.Duration
	ticker        *time.Ticker
	stopChan      chan bool
	stopOnce      sync.Once
}

// NewCache 创建新的缓存管理器
func NewCache(cacheDuration time.Duration) *Cache {
	cache := &Cache{
		cacheDuration: cacheDuration,
		stopChan:      make(chan bool),
	}

	// 启动定时清理过期缓存的任务
	cache.startCleanupTask()

	return cache
}

// Get 从缓存获取股票数据
func (c *Cache) Get(symbol string) (*models.StockData, bool) {
	if entry, exists := c.entries.Load(symbol); exists {
		cacheEntry := entry.(*CacheEntry)

		// 检查是否过期
		if time.Now().Before(cacheEntry.ExpiresAt) {
			return cacheEntry.Data, true
		}

		// 过期，删除缓存
		c.entries.Delete(symbol)
	}

	return nil, false
}

// Set 将股票数据存入缓存
func (c *Cache) Set(symbol string, data *models.StockData) {
	cacheEntry := &CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(c.cacheDuration),
	}

	c.entries.Store(symbol, cacheEntry)
}

// Delete 从缓存删除指定股票数据
func (c *Cache) Delete(symbol string) {
	c.entries.Delete(symbol)
}

// Clear 清空所有缓存
func (c *Cache) Clear() {
	c.entries.Range(func(key, value interface{}) bool {
		c.entries.Delete(key)
		return true
	})
}

// Size 获取缓存条目数量
func (c *Cache) Size() int {
	count := 0
	c.entries.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// GetExpiredCount 获取过期但未清理的缓存条目数量
func (c *Cache) GetExpiredCount() int {
	count := 0
	now := time.Now()

	c.entries.Range(func(key, value interface{}) bool {
		if entry, ok := value.(*CacheEntry); ok {
			if now.After(entry.ExpiresAt) {
				count++
			}
		}
		return true
	})

	return count
}

// Cleanup 清理过期的缓存条目
func (c *Cache) Cleanup() {
	now := time.Now()

	c.entries.Range(func(key, value interface{}) bool {
		if entry, ok := value.(*CacheEntry); ok {
			if now.After(entry.ExpiresAt) {
				c.entries.Delete(key)
			}
		}
		return true
	})
}

// startCleanupTask 启动定时清理任务
func (c *Cache) startCleanupTask() {
	// 每分钟清理一次过期缓存
	c.ticker = time.NewTicker(time.Minute)

	go func() {
		for {
			select {
			case <-c.ticker.C:
				c.Cleanup()
			case <-c.stopChan:
				c.ticker.Stop()
				return
			}
		}
	}()
}

// Stop 停止缓存管理器
func (c *Cache) Stop() {
	c.stopOnce.Do(func() {
		if c.stopChan != nil {
			close(c.stopChan)
		}
		if c.ticker != nil {
			c.ticker.Stop()
		}
	})

	c.Clear()
}

// SetCacheDuration 设置缓存时长
func (c *Cache) SetCacheDuration(duration time.Duration) {
	c.cacheDuration = duration
}

// GetCacheDuration 获取缓存时长
func (c *Cache) GetCacheDuration() time.Duration {
	return c.cacheDuration
}

// GetCacheStats 获取缓存统计信息
func (c *Cache) GetCacheStats() CacheStats {
	stats := CacheStats{
		TotalEntries:   c.Size(),
		ExpiredEntries: c.GetExpiredCount(),
	}

	return stats
}

// CacheStats 缓存统计信息
type CacheStats struct {
	TotalEntries   int `json:"total_entries"`
	ExpiredEntries int `json:"expired_entries"`
}

// String 返回统计信息的字符串表示
func (s CacheStats) String() string {
	return fmt.Sprintf("总条目数: %d, 过期条目数: %d", s.TotalEntries, s.ExpiredEntries)
}

// GetSymbolFromCacheEntry 从缓存条目获取股票代码
func GetSymbolFromCacheEntry(entry interface{}) string {
	if cacheEntry, ok := entry.(*CacheEntry); ok && cacheEntry.Data != nil {
		return cacheEntry.Data.Symbol
	}
	return ""
}

// IsCacheEntryValid 检查缓存条目是否有效
func IsCacheEntryValid(entry interface{}) bool {
	if cacheEntry, ok := entry.(*CacheEntry); ok {
		return time.Now().Before(cacheEntry.ExpiresAt) && cacheEntry.Data != nil
	}
	return false
}
