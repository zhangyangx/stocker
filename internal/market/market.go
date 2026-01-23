package market

import (
	"strings"
)

const (
	MarketCN = "CN"
)

// NormalizeSymbol 标准化输入股票代码，仅支持A股/场内ETF/指数。
// 允许输入：
// - 带前缀：sh600000 / sz000001
// - 纯数字：600000 / 000001（默认按A股规则补前缀）
func NormalizeSymbol(input string) (string, bool) {
	raw := strings.TrimSpace(strings.ToLower(input))
	if raw == "" {
		return "", false
	}

	// 已带前缀
	if strings.HasPrefix(raw, "sh") || strings.HasPrefix(raw, "sz") {
		code := strings.TrimPrefix(strings.TrimPrefix(raw, "sh"), "sz")
		if len(code) != 6 || !isAllDigits(code) {
			return "", false
		}
		if strings.HasPrefix(raw, "sh") {
			return "sh" + code, true
		}
		return "sz" + code, true
	}

	// 纯数字
	if len(raw) == 6 && isAllDigits(raw) {
		if raw[0] == '6' {
			return "sh" + raw, true
		}
		return "sz" + raw, true
	}

	return "", false
}

// DisplaySymbol 返回展示用的纯代码（不含前缀）。
func DisplaySymbol(symbol string) string {
	if len(symbol) >= 2 && (strings.HasPrefix(symbol, "sh") || strings.HasPrefix(symbol, "sz")) {
		return symbol[2:]
	}
	return symbol
}

// ExtractCode 提取6位纯代码。
func ExtractCode(symbol string) (string, bool) {
	if len(symbol) >= 2 && (strings.HasPrefix(symbol, "sh") || strings.HasPrefix(symbol, "sz")) {
		code := symbol[2:]
		if len(code) == 6 && isAllDigits(code) {
			return code, true
		}
		return "", false
	}
	if len(symbol) == 6 && isAllDigits(symbol) {
		return symbol, true
	}
	return "", false
}

// IsETF 判断是否为场内ETF（按代码段识别）。
// 说明：代码段可能随市场变化，必要时可扩展。
func IsETF(symbol string) bool {
	code, ok := ExtractCode(strings.ToLower(symbol))
	if !ok {
		return false
	}

	// 常见ETF代码段：沪市 510/511/512/513/515/516/517/518/519/588 等
	// 深市 159/160/161 等
	prefix3 := code[:3]
	switch prefix3 {
	case "510", "511", "512", "513", "515", "516", "517", "518", "519", "588",
		"159", "160", "161":
		return true
	default:
		return false
	}
}

// IsIndex 判断是否为常见指数代码（粗粒度）。
func IsIndex(symbol string) bool {
	code, ok := ExtractCode(strings.ToLower(symbol))
	if !ok {
		return false
	}
	if strings.HasPrefix(symbol, "sh") && strings.HasPrefix(code, "000") {
		return true
	}
	if strings.HasPrefix(symbol, "sz") && strings.HasPrefix(code, "399") {
		return true
	}
	return false
}

func isAllDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
