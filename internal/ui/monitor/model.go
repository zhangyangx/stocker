package monitor

import (
	"fmt"
	"stocker/internal/config"
	"stocker/internal/stock"
	"stocker/pkg/models"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// MonitorState 监控状态
type MonitorState int

const (
	StateRunning  MonitorState = iota // 运行中
	StatePaused                       // 已暂停
	StateSettings                     // 设置界面
	StateHelp                         // 帮助界面
)

// Model 实时监控模型
type Model struct {
	manager         *stock.Manager
	config          *config.Config
	state           MonitorState
	stockData       []*models.StockData
	lastUpdate      time.Time
	refreshInterval time.Duration
	err             error
	width           int
	height          int
	paused          bool
	pauseReason     PauseReason
	showHelp        bool
	settings        *SettingsModel
	connectionState ConnectionState
	statusMessage   string
	simpleMode      bool // 简洁模式标志
}

// PauseReason 暂停原因
type PauseReason int

const (
	PauseNone PauseReason = iota
	PauseManual
	PauseSchedule
)

// ConnectionState 连接状态
type ConnectionState int

const (
	ConnectionNormal ConnectionState = iota // 🟢 正常连接
	ConnectionSlow                          // 🟡 连接缓慢
	ConnectionFailed                        // 🔴 连接失败
)

// NewModel 创建新的监控模型
func NewModel(manager *stock.Manager, cfg *config.Config, simpleMode bool) Model {
	interval := time.Duration(cfg.Preferences.RefreshInterval) * time.Second
	if interval < time.Second {
		interval = 3 * time.Second
	}

	m := Model{
		manager:         manager,
		config:          cfg,
		state:           StateRunning,
		refreshInterval: interval,
		width:           80,
		height:          24,
		paused:          false,
		pauseReason:     PauseNone,
		showHelp:        false,
		settings:        NewSettingsModel(cfg),
		connectionState: ConnectionNormal,
		simpleMode:      simpleMode,
	}

	if paused, reason := shouldAutoPauseRefreshAt(time.Now()); paused {
		m.setPause(PauseSchedule, reason)
	}

	return m
}

// Init 初始化模型
func (m Model) Init() tea.Cmd {
	if m.pauseReason == PauseSchedule {
		return tea.Batch(
			tickCmd(m.refreshInterval),
			tea.EnterAltScreen,
		)
	}
	return tea.Batch(
		tickCmd(m.refreshInterval),
		fetchStockDataCmd(m.manager),
		tea.EnterAltScreen,
	)
}

// Update 更新模型
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// 如果在设置界面
		if m.state == StateSettings {
			return m.handleSettingsInput(msg)
		}

		// 如果在帮助界面
		if m.state == StateHelp {
			if msg.String() == "h" || msg.String() == "q" || msg.String() == "esc" {
				m.state = StateRunning
			}
			return m, nil
		}

		// 主界面键盘处理
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit

		case " ": // 空格键暂停/继续
			if m.pauseReason == PauseSchedule {
				m.statusMessage = "当前时段已自动暂停，可按 r 手动刷新"
				return m, nil
			}
			if m.pauseReason == PauseManual {
				m.setPause(PauseNone, "监控已继续")
				return m, fetchStockDataCmd(m.manager)
			}
			m.setPause(PauseManual, "监控已暂停")
			return m, nil

		case "r": // 手动刷新
			m.statusMessage = "正在刷新数据..."
			if m.manager != nil {
				m.manager.ClearCache()
			}
			return m, fetchStockDataCmd(m.manager)

		case "h": // 显示帮助
			m.state = StateHelp
			return m, nil

		case "s": // 进入设置
			m.state = StateSettings
			return m, nil
		}

	case tickMsg:
		if paused, reason := shouldAutoPauseRefreshAt(time.Time(msg)); paused {
			m.setPause(PauseSchedule, reason)
			return m, tickCmd(m.refreshInterval)
		}
		if m.pauseReason == PauseSchedule {
			m.setPause(PauseNone, "已恢复自动刷新")
		}
		// 手动暂停时，不自动刷新
		if m.pauseReason == PauseManual {
			return m, tickCmd(m.refreshInterval)
		}
		return m, tea.Batch(
			tickCmd(m.refreshInterval),
			fetchStockDataCmd(m.manager),
		)

	case stockDataMsg:
		m.err = msg.err
		if msg.err != nil {
			if len(m.stockData) > 0 {
				m.connectionState = ConnectionSlow
				m.statusMessage = fmt.Sprintf("刷新失败，展示最近一次成功数据: %v", msg.err)
			} else {
				m.stockData = msg.data
				m.connectionState = ConnectionFailed
				m.statusMessage = fmt.Sprintf("获取数据失败: %v", msg.err)
			}
		} else {
			m.stockData = msg.data
			m.lastUpdate = time.Now()
			m.connectionState = ConnectionNormal
			m.statusMessage = ""
		}
		return m, nil

	case errMsg:
		m.err = msg.err
		m.connectionState = ConnectionFailed
		return m, nil
	}

	return m, nil
}

func (m *Model) setPause(reason PauseReason, message string) {
	m.pauseReason = reason
	m.paused = reason != PauseNone
	if message != "" {
		m.statusMessage = message
	}
}

func shouldAutoPauseRefreshAt(at time.Time) (bool, string) {
	minutes := at.Hour()*60 + at.Minute()
	if minutes >= 11*60+35 && minutes < 13*60 {
		return true, "午间休市，已自动暂停刷新"
	}
	if minutes >= 15*60+5 {
		return true, "已收盘，自动暂停刷新"
	}
	return false, ""
}

// handleSettingsInput 处理设置界面的输入
func (m Model) handleSettingsInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.state = StateRunning
		return m, nil

	case "up", "k": // 向上移动选项
		if m.settings.selectedItem > 0 {
			m.settings.selectedItem--
		}
		return m, nil

	case "down", "j": // 向下移动选项
		if m.settings.selectedItem < 4 { // 共5个选项，索引0-4
			m.settings.selectedItem++
		}
		return m, nil

	case "1": // 修改刷新间隔
		m.settings.selectedItem = 0
		return m, nil

	case "2": // 修改显示列
		m.settings.selectedItem = 1
		return m, nil

	case "3": // 修改颜色主题
		m.settings.selectedItem = 2
		return m, nil

	case "4": // 修改排序方式
		m.settings.selectedItem = 3
		return m, nil

	case "5": // 修改简洁模式
		m.settings.selectedItem = 4
		return m, nil

	case "+", "=": // 增加刷新间隔
		if m.settings.selectedItem == 0 {
			interval := m.config.Preferences.RefreshInterval + 1
			if interval <= 10 {
				m.config.Preferences.RefreshInterval = interval
				m.refreshInterval = time.Duration(interval) * time.Second
				m.config.Save()
			}
		}
		return m, nil

	case "-": // 减少刷新间隔
		if m.settings.selectedItem == 0 {
			interval := m.config.Preferences.RefreshInterval - 1
			if interval >= 1 {
				m.config.Preferences.RefreshInterval = interval
				m.refreshInterval = time.Duration(interval) * time.Second
				m.config.Save()
			}
		}
		return m, nil

	case " ": // 空格切换简洁模式
		if m.settings.selectedItem == 4 {
			m.config.Preferences.SimpleMode = !m.config.Preferences.SimpleMode
			m.simpleMode = m.config.Preferences.SimpleMode
			m.config.Save()
		}
		return m, nil

	case "enter":
		// 保存设置并返回
		m.state = StateRunning
		m.statusMessage = "设置已保存"
		return m, nil
	}

	return m, nil
}

// View 渲染视图
func (m Model) View() string {
	switch m.state {
	case StateSettings:
		return m.renderSettings()
	case StateHelp:
		return m.renderHelp()
	default:
		return m.renderMonitor()
	}
}
