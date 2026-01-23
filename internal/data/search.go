package data

import (
	"fmt"
	"strconv"
	"strings"

	"stocker/internal/config"
	"stocker/internal/market"
	"stocker/pkg/models"
)

// SearchStock 根据关键词搜索股票（使用新浪API）
func (c *SinaClient) SearchStock(keyword string) ([]models.SearchStock, error) {
	// 首先尝试新浪搜索
	results, err := c.searchFromSina(keyword)
	if err != nil {
		// 新浪搜索失败，尝试腾讯搜索
		tencentResults, tencentErr := c.searchFromTencent(keyword)
		if tencentErr != nil && err != nil {
			return nil, fmt.Errorf("新浪搜索失败: %v, 腾讯搜索失败: %v", err, tencentErr)
		}
		if tencentErr != nil {
			return nil, err
		}
		return tencentResults, nil
	}

	// 如果新浪搜索结果不足，补充腾讯搜索结果
	if len(results) < 5 {
		tencentResults, tencentErr := c.searchFromTencent(keyword)
		if tencentErr == nil {
			// 合并结果，去重
			results = mergeSearchResults(results, tencentResults)
		}
	}

	return results[:min(len(results), 10)], nil // 最多返回10个结果
}

// searchFromSina 从新浪API搜索股票
func (c *SinaClient) searchFromSina(keyword string) ([]models.SearchStock, error) {
	url := config.SinaSearchURL + keyword
	headers := map[string]string{
		"Referer": config.SinaReferer,
	}

	response, err := c.request(url, headers)
	if err != nil {
		return nil, fmt.Errorf("新浪搜索API请求失败: %w", err)
	}

	return parseSinaSearchResult(response)
}

// searchFromTencent 从腾讯API搜索股票
func (c *SinaClient) searchFromTencent(keyword string) ([]models.SearchStock, error) {
	url := config.TencentSearchURL + keyword

	response, err := c.request(url, nil)
	if err != nil {
		return nil, fmt.Errorf("腾讯搜索API请求失败: %w", err)
	}

	return parseTencentSearchResult(response)
}

// parseSinaSearchResult 解析新浪搜索结果
func parseSinaSearchResult(response string) ([]models.SearchStock, error) {
	// 查找suggestvalue
	if !strings.Contains(response, "suggestvalue=") {
		return nil, fmt.Errorf("搜索结果格式错误")
	}

	dataStart := strings.Index(response, "suggestvalue=\"") + len("suggestvalue=\"")
	dataEnd := strings.Index(response[dataStart:], "\"")
	if dataEnd == -1 {
		return nil, fmt.Errorf("搜索结果格式错误")
	}

	dataStr := response[dataStart : dataStart+dataEnd]
	if dataStr == "" {
		return []models.SearchStock{}, nil
	}

	stocks := strings.Split(dataStr, ";")
	var results []models.SearchStock

	for _, stock := range stocks {
		if stock == "" {
			continue
		}

		fields := strings.Split(stock, ",")
		if len(fields) < 5 {
			continue
		}

		code := fields[0]
		typeID := fields[1]
		name := decodeUnicodeEscapes(fields[4])

		// 根据类型ID过滤结果并格式化代码
		formattedCode, ok := formatSinaSearchCode(typeID, code)
		if !ok {
			continue
		}

		// 过滤掉ST股票等
		if strings.Contains(name, "S*ST") || strings.Contains(name, "*ST") {
			continue
		}

		// 仅保留A股/ETF/指数
		if !isAllowedAStock(formattedCode) && !market.IsETF(formattedCode) && !market.IsIndex(formattedCode) {
			continue
		}

		results = append(results, models.SearchStock{
			Code:   formattedCode,
			Name:   name,
			Market: string(models.MarketCN),
		})

		// 限制结果数量
		if len(results) >= 10 {
			break
		}
	}

	return results, nil
}

// parseTencentSearchResult 解析腾讯搜索结果
func parseTencentSearchResult(response string) ([]models.SearchStock, error) {
	// 查找v_hint
	if !strings.Contains(response, "v_hint=") {
		return nil, fmt.Errorf("搜索结果格式错误")
	}

	dataStart := strings.Index(response, "v_hint=\"") + len("v_hint=\"")
	dataEnd := strings.Index(response[dataStart:], "\"")
	if dataEnd == -1 {
		return nil, fmt.Errorf("搜索结果格式错误")
	}

	dataStr := response[dataStart : dataStart+dataEnd]
	if dataStr == "" {
		return []models.SearchStock{}, nil
	}

	stocks := strings.Split(dataStr, "^")
	var results []models.SearchStock

	for _, stock := range stocks {
		if stock == "" {
			continue
		}

		fields := strings.Split(stock, "~")
		if len(fields) < 3 {
			continue
		}

		marketCode := fields[0]
		code := fields[1]
		name := decodeUnicodeEscapes(fields[2])

		// 根据市场代码转换（仅A股）
		formattedCode, ok := parseTencentSearchType(marketCode, code)
		if !ok {
			continue
		}

		if !isAllowedAStock(formattedCode) && !market.IsETF(formattedCode) && !market.IsIndex(formattedCode) {
			continue
		}

		results = append(results, models.SearchStock{
			Code:   formattedCode,
			Name:   name,
			Market: string(models.MarketCN),
		})

		// 限制结果数量
		if len(results) >= 10 {
			break
		}
	}

	return results, nil
}

// formatSinaSearchCode 根据类型ID格式化股票代码
func formatSinaSearchCode(typeID, code string) (formattedCode string, ok bool) {
	// 如果代码已有前缀，直接使用
	if strings.HasPrefix(code, "sh") || strings.HasPrefix(code, "sz") {
		return code, true
	}

	switch typeID {
	case "11": // 上海A股
		return "sh" + code, true
	case "22", "81": // 深圳A股
		return "sz" + code, true
	default:
		return "", false
	}
}

// parseTencentSearchType 解析腾讯搜索结果市场代码
func parseTencentSearchType(marketCode, code string) (string, bool) {
	switch marketCode {
	case "sz":
		if len(code) == 6 {
			return "sz" + code, true
		}
		return code, true
	case "sh":
		if len(code) == 6 {
			return "sh" + code, true
		}
		return code, true
	default:
		return "", false
	}
}

// mergeSearchResults 合并搜索结果，去重
func mergeSearchResults(sinaResults, tencentResults []models.SearchStock) []models.SearchStock {
	seen := make(map[string]bool)
	var merged []models.SearchStock

	// 先添加新浪结果
	for _, result := range sinaResults {
		if !seen[result.Code] {
			seen[result.Code] = true
			merged = append(merged, result)
		}
	}

	// 再添加腾讯结果中不重复的
	for _, result := range tencentResults {
		if !seen[result.Code] {
			seen[result.Code] = true
			merged = append(merged, result)
		}
	}

	return merged
}

// FormatStockCode 格式化股票代码显示
func FormatStockCode(code string) string {
	// 去掉前缀显示
	if strings.HasPrefix(code, "sh") || strings.HasPrefix(code, "sz") {
		return code[2:] + " (A股)"
	}
	return code
}

func isAllowedAStock(symbol string) bool {
	norm, ok := market.NormalizeSymbol(symbol)
	if !ok {
		return false
	}
	return strings.HasPrefix(norm, "sh") || strings.HasPrefix(norm, "sz")
}

func decodeUnicodeEscapes(s string) string {
	if !strings.Contains(s, "\\u") {
		return s
	}
	escaped := strings.ReplaceAll(s, "\"", "\\\"")
	if decoded, err := strconv.Unquote("\"" + escaped + "\""); err == nil {
		return decoded
	}
	return s
}

// 求最小值的辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
