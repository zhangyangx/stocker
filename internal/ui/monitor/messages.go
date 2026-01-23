package monitor

import (
	"stocker/internal/stock"
	"stocker/pkg/models"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// tickMsg 定时器消息
type tickMsg time.Time

// tickCmd 创建定时器命令
func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// stockDataMsg 股票数据消息
type stockDataMsg struct {
	data []*models.StockData
	err  error
}

// fetchStockDataCmd 获取股票数据命令
func fetchStockDataCmd(manager *stock.Manager) tea.Cmd {
	return func() tea.Msg {
		data, err := manager.GetAllStockData()
		return stockDataMsg{
			data: data,
			err:  err,
		}
	}
}

// errMsg 错误消息
type errMsg struct {
	err error
}
