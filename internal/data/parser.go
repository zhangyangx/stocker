package data

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"stocker/internal/market"
	"stocker/pkg/models"
)

// parseSinaStockData 解析新浪API返回的股票数据
func parseSinaStockData(symbol, response string) (*models.StockData, error) {
	// 查找对应的数据行
	dataLine := findSinaDataLine(symbol, response)
	if dataLine == "" {
		return nil, fmt.Errorf("未找到股票 %s 的数据", symbol)
	}

	// 提取数据部分
	if !strings.Contains(dataLine, "=\"") {
		return nil, fmt.Errorf("数据格式错误")
	}

	dataStart := strings.Index(dataLine, "=\"") + 2
	dataEnd := strings.LastIndex(dataLine, "\"")
	if dataEnd <= dataStart {
		return nil, fmt.Errorf("数据格式错误")
	}

	dataStr := dataLine[dataStart:dataEnd]
	if dataStr == "" {
		return nil, fmt.Errorf("股票数据为空")
	}

	fields := strings.Split(dataStr, ",")
	if len(fields) < 10 {
		return nil, fmt.Errorf("数据字段不足")
	}

	// 仅支持A股/ETF/指数
	norm, ok := market.NormalizeSymbol(symbol)
	if !ok {
		return nil, fmt.Errorf("不支持的股票代码: %s", symbol)
	}

	stockData, err := parseSinaA股Data(norm, fields)
	if err != nil {
		return nil, err
	}

	// 设置统一的市场类型
	stockData.Market = string(models.MarketCN)
	stockData.Timestamp = time.Now()

	return stockData, nil
}

// findSinaDataLine 在响应中查找指定股票的数据行
func findSinaDataLine(symbol string, response string) string {
	pattern := "hq_str_" + symbol
	lines := strings.Split(response, ";\n")

	for _, line := range lines {
		if strings.Contains(line, pattern) {
			return line
		}
	}

	return ""
}

// parseSinaA股Data 解析新浪A股数据
func parseSinaA股Data(symbol string, fields []string) (*models.StockData, error) {
	// A股数据字段说明（32个字段，取前33个）
	// [0]名称 [1]开盘 [2]昨收 [3]当前价 [4]最高 [5]最低 [6]买入价 [7]卖出价
	// [8]成交量 [9]成交额 [31]日期 [32]时间

	if len(fields) < 10 {
		return nil, fmt.Errorf("A股数据字段不足")
	}

	// 修正索引映射
	name := strings.TrimSpace(fields[0]) // 名称在字段0
	opening := parseFloat(fields[1])     // 开盘价在字段1
	close := parseFloat(fields[2])       // 昨收价在字段2
	current := parseFloat(fields[3])     // 当前价在字段3
	high := parseFloat(fields[4])        // 最高价在字段4
	low := parseFloat(fields[5])         // 最低价在字段5
	volume := formatVolume(fields[8])    // 成交量在字段8

	// 计算涨跌额和涨跌幅
	change := current - close
	changePercent := 0.0
	if close > 0 {
		changePercent = change / close * 100
	}

	return &models.StockData{
		Symbol:        symbol,
		Name:          name,
		Current:       current,
		Opening:       opening,
		Close:         close,
		High:          high,
		Low:           low,
		Change:        change,
		ChangePercent: changePercent,
		Volume:        volume,
	}, nil
}

// parseSina港股Data 解析新浪港股数据
// parseTencentStockData 解析腾讯API返回的股票数据
func parseTencentStockData(symbol, response string) (*models.StockData, error) {
	// 查找对应的数据行
	dataLine := findTencentDataLine(symbol, response)
	if dataLine == "" {
		return nil, fmt.Errorf("未找到股票 %s 的数据", symbol)
	}

	// 检查是否有效
	if strings.Contains(dataLine, "pv_none_match") {
		return nil, fmt.Errorf("无效的股票代码: %s", symbol)
	}

	// 提取数据部分
	if !strings.Contains(dataLine, "=\"") {
		return nil, fmt.Errorf("数据格式错误")
	}

	dataStart := strings.Index(dataLine, "=\"") + 2
	dataEnd := strings.LastIndex(dataLine, "\"")
	if dataEnd <= dataStart {
		return nil, fmt.Errorf("数据格式错误")
	}

	dataStr := dataLine[dataStart:dataEnd]
	if dataStr == "" {
		return nil, fmt.Errorf("股票数据为空")
	}

	fields := strings.Split(dataStr, "~")
	if len(fields) < 35 {
		return nil, fmt.Errorf("数据字段不足")
	}

	// 解析腾讯数据（统一格式，36+字段，波浪号分隔）
	// [1]名称 [3]当前价 [4]昨收 [5]开盘 [6]最高 [7]最低 [31]时间戳 [33]涨跌幅(%)
	name := strings.TrimSpace(fields[1])
	current := parseFloat(fields[3])
	close := parseFloat(fields[4])
	opening := parseFloat(fields[5])
	high := parseFloat(fields[6])
	low := parseFloat(fields[7])
	changePercent := parseFloat(fields[33])

	// 计算涨跌额
	change := current - close

	// 解析时间戳
	timestampStr := fields[31]
	timestamp, _ := parseTencentTimestamp(timestampStr)

	return &models.StockData{
		Symbol:        symbol,
		Name:          name,
		Current:       current,
		Opening:       opening,
		Close:         close,
		High:          high,
		Low:           low,
		Change:        change,
		ChangePercent: changePercent,
		Volume:        "", // 腾讯API成交量字段位置需要确认
		Timestamp:     timestamp,
	}, nil
}

// findTencentDataLine 在响应中查找指定股票的数据行
func findTencentDataLine(symbol string, response string) string {
	pattern := "v_" + symbol + "="
	lines := strings.Split(response, ";\n")

	for _, line := range lines {
		if strings.Contains(line, pattern) {
			return line
		}
	}

	return ""
}

// parseFloat 安全地解析浮点数
func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "--" || s == "0.00" {
		return 0.0
	}

	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0
	}
	return value
}

// formatVolume 格式化成交量显示
func formatVolume(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	// 尝试转换为数字
	volume, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return s
	}

	// 转换为手（1手=股）
	if volume > 10000 {
		return fmt.Sprintf("%.1fK", volume/10000)
	} else if volume > 1000000 {
		return fmt.Sprintf("%.1fM", volume/1000000)
	}

	return fmt.Sprintf("%.0f", volume)
}

// parseTencentTimestamp 解析腾讯时间戳
func parseTencentTimestamp(timestampStr string) (time.Time, error) {
	// 腾讯时间戳格式: 20240123150302
	if len(timestampStr) != 14 {
		return time.Now(), fmt.Errorf("时间戳格式错误")
	}

	layout := "20060102150405"
	return time.Parse(layout, timestampStr)
}
