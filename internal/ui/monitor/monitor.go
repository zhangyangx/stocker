package monitor

import (
	"fmt"
	"stocker/internal/config"
	"stocker/internal/stock"

	tea "github.com/charmbracelet/bubbletea"
)

// Start 启动实时监控
func Start(manager *stock.Manager, cfg *config.Config, simpleMode bool) error {
	// 检查监控列表
	watchlist := manager.GetWatchlist()
	if len(watchlist) == 0 {
		return fmt.Errorf("监控列表为空，请先添加股票")
	}

	// 创建监控模型
	m := NewModel(manager, cfg, simpleMode)

	// 创建程序
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// 运行程序
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("运行监控失败: %w", err)
	}

	return nil
}
