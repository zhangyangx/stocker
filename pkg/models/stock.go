package models

import (
	"fmt"
	"strconv"
	"time"
)

// Stock 股票基础信息结构
type Stock struct {
	Symbol  string    `json:"symbol"`   // 股票代码（不含前缀）
	Name    string    `json:"name"`     // 股票名称
	Market  string    `json:"market"`   // 市场类型（CN）
	AddedAt time.Time `json:"added_at"` // 添加时间
}

// StockData 股票实时数据结构
type StockData struct {
	Symbol        string    `json:"symbol"`         // 股票代码（含前缀）
	Name          string    `json:"name"`           // 股票名称
	Market        string    `json:"market"`         // 市场类型
	Current       float64   `json:"price"`          // 当前价格
	Opening       float64   `json:"opening"`        // 开盘价
	Close         float64   `json:"close"`          // 昨收价
	High          float64   `json:"high"`           // 最高价
	Low           float64   `json:"low"`            // 最低价
	Change        float64   `json:"change"`         // 涨跌额
	ChangePercent float64   `json:"change_percent"` // 涨跌幅
	Volume        string    `json:"volume"`         // 成交量
	Timestamp     time.Time `json:"timestamp"`      // 更新时间
}

// SearchStock 搜索结果结构
type SearchStock struct {
	Code   string `json:"code"`   // 格式化后的股票代码
	Name   string `json:"name"`   // 股票名称
	Market string `json:"market"` // 市场类型
}

// Watchlist 监控列表结构
type Watchlist struct {
	Stocks   []Stock `json:"watchlist"` // 监控股票列表
	Settings struct {
		RefreshInterval int `json:"refresh_interval"` // 刷新间隔（秒）
	} `json:"settings"`
}

// MarketType 市场类型枚举
type MarketType string

const (
	MarketCN MarketType = "CN" // A股
)

// IsExpired 检查股票数据是否过期
func (s *StockData) IsExpired(cacheDuration time.Duration) bool {
	return time.Since(s.Timestamp) > cacheDuration
}

// GetDisplaySymbol 获取显示用的股票代码（根据市场决定显示格式）
func (s *StockData) GetDisplaySymbol() string {
	switch s.Market {
	case string(MarketCN):
		if len(s.Symbol) == 9 && (s.Symbol[:2] == "sh" || s.Symbol[:2] == "sz") {
			return s.Symbol[2:]
		}
		return s.Symbol
	default:
		return s.Symbol
	}
}

// GetMarketName 获取市场显示名称
func (s *StockData) GetMarketName() string {
	switch s.Market {
	case string(MarketCN):
		return "A股"
	default:
		return "未知"
	}
}

// GetChangeColor 获取涨跌幅对应的颜色标记（用于终端显示）
func (s *StockData) GetChangeColor() string {
	if s.ChangePercent > 0 {
		return "green" // 涨
	} else if s.ChangePercent < 0 {
		return "red" // 跌
	}
	return "white" // 平
}

// FormatPrice 输出价格的自然字符串表示，不额外截断或补零。
func FormatPrice(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

// FormatChange 格式化涨跌幅显示
func (s *StockData) FormatChange() string {
	if s.ChangePercent > 0 {
		return fmt.Sprintf("+%.2f%%", s.ChangePercent)
	} else if s.ChangePercent < 0 {
		return fmt.Sprintf("%.2f%%", s.ChangePercent)
	}
	return "0.00%"
}
