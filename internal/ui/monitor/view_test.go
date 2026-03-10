package monitor

import (
	"strings"
	"testing"

	"stocker/pkg/models"
)

func TestRenderTableRowUsesNaturalPriceString(t *testing.T) {
	m := Model{simpleMode: true}
	stock := &models.StockData{
		Symbol:        "sz300750",
		Name:          "宁德时代",
		Market:        string(models.MarketCN),
		Current:       12.345,
		ChangePercent: 1.23,
	}
	widths := ColumnWidths{
		Index:  3,
		Code:   8,
		Name:   8,
		Price:  8,
		Change: 8,
	}

	row := m.renderTableRow(1, stock, widths, false, false)

	if !strings.Contains(row, "12.345") {
		t.Fatalf("expected monitor row to contain natural price precision, got: %q", row)
	}
	if strings.Contains(row, "12.35") {
		t.Fatalf("expected monitor row to avoid rounding to two decimals, got: %q", row)
	}
}
