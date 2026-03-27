package monitor

import (
	"errors"
	"strings"
	"testing"
	"time"

	"stocker/internal/config"
	"stocker/pkg/models"
)

func TestStockDataErrorKeepsPreviousDisplayAndMarksSlow(t *testing.T) {
	m := NewModel(nil, config.GetDefaultConfig(), true)
	previous := []*models.StockData{
		{
			Symbol:        "sz300750",
			Name:          "宁德时代",
			Market:        string(models.MarketCN),
			Current:       250.12,
			ChangePercent: 1.25,
			Timestamp:     time.Date(2026, 3, 27, 10, 0, 0, 0, time.Local),
		},
	}
	m.stockData = previous
	m.lastUpdate = previous[0].Timestamp
	m.connectionState = ConnectionNormal

	next, _ := m.Update(stockDataMsg{err: errors.New("network timeout")})
	updated := next.(Model)

	if len(updated.stockData) != 1 || updated.stockData[0].Symbol != previous[0].Symbol {
		t.Fatalf("expected previous stock data to be retained, got %#v", updated.stockData)
	}
	if updated.connectionState != ConnectionSlow {
		t.Fatalf("expected connection state slow, got %v", updated.connectionState)
	}
	if !strings.Contains(updated.statusMessage, "最近一次成功数据") {
		t.Fatalf("expected stale-data status message, got %q", updated.statusMessage)
	}
	if updated.lastUpdate != previous[0].Timestamp {
		t.Fatalf("expected lastUpdate to stay unchanged, got %v", updated.lastUpdate)
	}
}

func TestStockDataErrorWithoutPreviousDataMarksFailed(t *testing.T) {
	m := NewModel(nil, config.GetDefaultConfig(), true)

	next, _ := m.Update(stockDataMsg{err: errors.New("provider unavailable")})
	updated := next.(Model)

	if updated.connectionState != ConnectionFailed {
		t.Fatalf("expected connection state failed, got %v", updated.connectionState)
	}
	if !strings.Contains(updated.statusMessage, "获取数据失败") {
		t.Fatalf("expected failure status message, got %q", updated.statusMessage)
	}
}

func TestShouldAutoPauseRefreshAt(t *testing.T) {
	cases := []struct {
		name   string
		at     time.Time
		paused bool
	}{
		{
			name:   "before lunch remains active",
			at:     time.Date(2026, 3, 27, 11, 34, 0, 0, time.Local),
			paused: false,
		},
		{
			name:   "lunch break pauses",
			at:     time.Date(2026, 3, 27, 11, 35, 0, 0, time.Local),
			paused: true,
		},
		{
			name:   "afternoon session resumes",
			at:     time.Date(2026, 3, 27, 13, 0, 0, 0, time.Local),
			paused: false,
		},
		{
			name:   "after close pauses",
			at:     time.Date(2026, 3, 27, 15, 5, 0, 0, time.Local),
			paused: true,
		},
		{
			name:   "pre-market is not auto paused",
			at:     time.Date(2026, 3, 27, 9, 0, 0, 0, time.Local),
			paused: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			paused, _ := shouldAutoPauseRefreshAt(tc.at)
			if paused != tc.paused {
				t.Fatalf("expected paused=%v, got %v", tc.paused, paused)
			}
		})
	}
}
