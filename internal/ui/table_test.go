package ui

import (
	"strings"
	"testing"
	"time"

	"stocker/pkg/models"
)

func TestDisplayStockListUsesNaturalPriceString(t *testing.T) {
	stocks := []*models.StockData{
		{
			Symbol:        "sz300750",
			Name:          "宁德时代",
			Market:        string(models.MarketCN),
			Current:       12.345,
			ChangePercent: 1.23,
		},
	}

	output := DisplayStockList(stocks, nil)

	if !strings.Contains(output, "12.345") {
		t.Fatalf("expected stock list to contain natural price precision, got: %q", output)
	}
	if strings.Contains(output, "12.35") {
		t.Fatalf("expected stock list to avoid rounding to two decimals, got: %q", output)
	}
}

func TestDisplaySingleStockUsesNaturalPriceString(t *testing.T) {
	stock := &models.StockData{
		Symbol:        "sz300750",
		Name:          "宁德时代",
		Market:        string(models.MarketCN),
		Current:       12.345,
		Opening:       12.301,
		Close:         12.299,
		High:          12.399,
		Low:           12.111,
		Change:        0.046,
		ChangePercent: 0.37,
		Timestamp:     time.Date(2026, 3, 10, 9, 30, 0, 0, time.UTC),
	}

	output := DisplaySingleStock(stock)

	for _, want := range []string{"最新价格: 12.345", "开盘价格: 12.301", "昨收价格: 12.299", "最高价格: 12.399", "最低价格: 12.111"} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected single stock output to contain %q, got: %q", want, output)
		}
	}
	if strings.Contains(output, "最新价格: 12.35") {
		t.Fatalf("expected single stock output to avoid rounding current price, got: %q", output)
	}
}
