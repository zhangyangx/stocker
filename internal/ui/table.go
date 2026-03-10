package ui

import (
	"fmt"
	"strings"
	"unicode"

	"stocker/internal/data"
	"stocker/internal/market"
	"stocker/pkg/models"
)

// displayWidth 计算字符串在终端中的显示宽度
// 中文字符占2个宽度，ASCII字符占1个宽度
func displayWidth(s string) int {
	width := 0
	for _, r := range s {
		if r > unicode.MaxASCII {
			width += 2 // 中文等宽字符
		} else {
			width += 1
		}
	}
	return width
}

// padRight 右填充空格，使字符串达到指定的显示宽度（左对齐）
func padRight(s string, width int) string {
	currentWidth := displayWidth(s)
	if currentWidth >= width {
		return s
	}
	return s + strings.Repeat(" ", width-currentWidth)
}

// padLeft 左填充空格，使字符串达到指定的显示宽度（右对齐）
func padLeft(s string, width int) string {
	currentWidth := displayWidth(s)
	if currentWidth >= width {
		return s
	}
	return strings.Repeat(" ", width-currentWidth) + s
}

// truncateToWidth 截断字符串到指定的显示宽度
func truncateToWidth(s string, maxWidth int) string {
	currentWidth := 0
	result := ""
	for _, r := range s {
		rWidth := 1
		if r > unicode.MaxASCII {
			rWidth = 2
		}
		if currentWidth+rWidth > maxWidth-1 { // 留1个位置给省略号
			return result + "…"
		}
		result += string(r)
		currentWidth += rWidth
	}
	return s
}

// TableConfig 表格配置
type TableConfig struct {
	ShowIndex      bool   // 是否显示序号
	ShowMarket     bool   // 是否显示市场
	IndexLabel     string // 序号列标题
	MaxNameLength  int    // 名称最大显示长度
	FloatPrecision int    // 浮点数精度
}

// DefaultTableConfig 默认表格配置
func DefaultTableConfig() *TableConfig {
	return &TableConfig{
		ShowIndex:      true,
		ShowMarket:     false,
		IndexLabel:     "序号",
		MaxNameLength:  12,
		FloatPrecision: 2,
	}
}

// DisplayWatchlistLocal 显示监控列表（仅本地配置字段，不拉取实时行情）
func DisplayWatchlistLocal(stocks []models.Stock, config *TableConfig) string {
	if config == nil {
		config = DefaultTableConfig()
	}

	if len(stocks) == 0 {
		return "当前监控股票列表为空"
	}

	// 定义列宽
	const (
		colIndex = 4
		colCode  = 12
		colName  = 16
	)

	var builder strings.Builder

	// 标题
	title := fmt.Sprintf("当前监控股票列表 (共 %d 只):", len(stocks))
	builder.WriteString(title + "\n")
	builder.WriteString(strings.Repeat("─", colIndex+colCode+colName+2) + "\n\n")

	// 表头 - 使用对齐函数
	builder.WriteString(padRight("序号", colIndex) + " " + padRight("代码", colCode) + " " + padRight("名称", colName) + "\n")
	builder.WriteString(strings.Repeat("─", colIndex) + " " + strings.Repeat("─", colCode) + " " + strings.Repeat("─", colName) + "\n")

	for i, stock := range stocks {
		code := getLocalDisplaySymbol(stock)
		name := stock.Name

		// 处理名称过长
		if displayWidth(name) > colName {
			name = truncateToWidth(name, colName)
		}

		// 使用对齐函数输出
		indexStr := fmt.Sprintf("%d", i+1)
		builder.WriteString(padRight(indexStr, colIndex) + " " + padRight(code, colCode) + " " + padRight(name, colName) + "\n")
	}

	// 底部分隔线
	builder.WriteString(strings.Repeat("─", colIndex) + " " + strings.Repeat("─", colCode) + " " + strings.Repeat("─", colName) + "\n")

	return builder.String()
}

func getLocalDisplaySymbol(stock models.Stock) string {
	// 目标：显示“纯代码/美股代码”，不混入市场标记（市场列单独显示）
	return market.DisplaySymbol(stock.Symbol)
}

// DisplayStockList 显示股票列表表格
func DisplayStockList(stocks []*models.StockData, config *TableConfig) string {
	if config == nil {
		config = DefaultTableConfig()
	}

	if len(stocks) == 0 {
		return "当前监控股票列表为空"
	}

	// 定义列宽
	const (
		colIndex  = 4
		colCode   = 12
		colName   = 16
		colPrice  = 12
		colChange = 10
	)

	var builder strings.Builder

	// 标题
	title := fmt.Sprintf("当前监控股票列表 (共 %d 只):", len(stocks))
	builder.WriteString(title + "\n")
	builder.WriteString(strings.Repeat("─", colIndex+colCode+colName+colPrice+colChange+4) + "\n\n")

	// 表头 - 使用对齐函数
	builder.WriteString(padRight("序号", colIndex) + " " +
		padRight("代码", colCode) + " " +
		padRight("名称", colName) + " " +
		padLeft("当前价格", colPrice) + " " +
		padLeft("涨跌幅", colChange) + "\n")
	builder.WriteString(strings.Repeat("─", colIndex) + " " +
		strings.Repeat("─", colCode) + " " +
		strings.Repeat("─", colName) + " " +
		strings.Repeat("─", colPrice) + " " +
		strings.Repeat("─", colChange) + "\n")

	// 数据行
	for i, stock := range stocks {
		code := stock.GetDisplaySymbol()
		name := stock.Name
		change := stock.FormatChange()

		// 处理名称过长
		if displayWidth(name) > colName {
			name = truncateToWidth(name, colName)
		}

		// 使用对齐函数输出
		indexStr := fmt.Sprintf("%d", i+1)
		priceStr := models.FormatPrice(stock.Current)

		builder.WriteString(padRight(indexStr, colIndex) + " " +
			padRight(code, colCode) + " " +
			padRight(name, colName) + " " +
			padLeft(priceStr, colPrice) + " " +
			padLeft(change, colChange) + "\n")
	}

	// 底部分隔线
	builder.WriteString(strings.Repeat("─", colIndex) + " " +
		strings.Repeat("─", colCode) + " " +
		strings.Repeat("─", colName) + " " +
		strings.Repeat("─", colPrice) + " " +
		strings.Repeat("─", colChange) + "\n")

	return builder.String()
}

// buildHeaders 构建表头（用于调试）
func buildHeaders(config *TableConfig) []string {
	var headers []string
	return headers
}

// DisplaySearchResults 显示搜索结果
func DisplaySearchResults(results []models.SearchStock, keyword string) string {
	if len(results) == 0 {
		return fmt.Sprintf("未找到包含关键词 '%s' 的股票", keyword)
	}

	// 定义列宽
	const (
		colIndex  = 4
		colCode   = 10
		colName   = 16
		colMarket = 6
	)

	var builder strings.Builder

	// 标题
	title := fmt.Sprintf("搜索 '%s' 的结果 (共 %d 个):", keyword, len(results))
	builder.WriteString(title + "\n")
	builder.WriteString(strings.Repeat("=", colIndex+colCode+colName+colMarket+3) + "\n")

	// 表头
	builder.WriteString(padRight("序号", colIndex) + " " +
		padRight("代码", colCode) + " " +
		padRight("名称", colName) + " " +
		padRight("市场", colMarket) + "\n")
	builder.WriteString(strings.Repeat("-", colIndex) + " " +
		strings.Repeat("-", colCode) + " " +
		strings.Repeat("-", colName) + " " +
		strings.Repeat("-", colMarket) + "\n")

	// 数据行
	for i, result := range results {
		codeDisplay := data.FormatStockCode(result.Code)
		name := result.Name

		// 处理名称过长
		if displayWidth(name) > colName {
			name = truncateToWidth(name, colName)
		}

		indexStr := fmt.Sprintf("%d", i+1)
		market := getMarketDisplayName(result.Market)

		builder.WriteString(padRight(indexStr, colIndex) + " " +
			padRight(codeDisplay, colCode) + " " +
			padRight(name, colName) + " " +
			padRight(market, colMarket) + "\n")
	}

	// 底部
	builder.WriteString(strings.Repeat("=", colIndex+colCode+colName+colMarket+3))
	builder.WriteString("\n")

	return builder.String()
}

// DisplaySingleStock 显示单只股票的详细信息
func DisplaySingleStock(stock *models.StockData) string {
	if stock == nil {
		return "股票数据为空"
	}

	var builder strings.Builder

	// 标题
	title := fmt.Sprintf("%s (%s)", stock.Name, stock.GetDisplaySymbol())
	builder.WriteString(title + "\n")
	builder.WriteString(strings.Repeat("─", len(title)) + "\n\n")

	// 基本信息
	builder.WriteString(fmt.Sprintf("市场: %s\n", stock.GetMarketName()))
	builder.WriteString(fmt.Sprintf("最新价格: %s\n", models.FormatPrice(stock.Current)))
	builder.WriteString(fmt.Sprintf("开盘价格: %s\n", models.FormatPrice(stock.Opening)))
	builder.WriteString(fmt.Sprintf("昨收价格: %s\n", models.FormatPrice(stock.Close)))
	builder.WriteString(fmt.Sprintf("最高价格: %s\n", models.FormatPrice(stock.High)))
	builder.WriteString(fmt.Sprintf("最低价格: %s\n", models.FormatPrice(stock.Low)))
	builder.WriteString(fmt.Sprintf("涨跌金额: %+.2f\n", stock.Change))
	builder.WriteString(fmt.Sprintf("涨跌幅度: %s\n", stock.FormatChange()))

	if stock.Volume != "" {
		builder.WriteString(fmt.Sprintf("成交量: %s\n", stock.Volume))
	}

	builder.WriteString(fmt.Sprintf("更新时间: %s\n", stock.Timestamp.Format("2006-01-02 15:04:05")))

	return builder.String()
}

// getMarketDisplayName 获取市场显示名称
func getMarketDisplayName(market string) string {
	switch market {
	case string(models.MarketCN):
		return "A股"
	default:
		return "未知"
	}
}
