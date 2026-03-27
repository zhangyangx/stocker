package monitor

import (
	"fmt"
	"strings"

	"stocker/pkg/models"

	"github.com/charmbracelet/lipgloss"
)

var (
	// 颜色样式
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00D9FF")).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#6C5CE7")).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B")).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFB86C")).
			Bold(true)

	// 涨跌颜色
	upStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50FA7B")).
		Bold(true)

	downStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
			Bold(true)

	neutralStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2"))
)

// renderMonitor 渲染主监控界面
func (m Model) renderMonitor() string {
	var b strings.Builder

	// 标题栏
	b.WriteString(m.renderTitleBar())
	b.WriteString("\n\n")

	// 如果没有数据
	if len(m.stockData) == 0 {
		if m.err != nil {
			if m.simpleMode {
				b.WriteString(fmt.Sprintf("错误: %v", m.err))
			} else {
				b.WriteString(errorStyle.Render(fmt.Sprintf("错误: %v", m.err)))
			}
		} else {
			b.WriteString("当前监控列表为空，请先添加股票到监控列表\n")
			b.WriteString("使用 'search <关键词>' 命令搜索并添加股票")
		}
		b.WriteString("\n\n")
		b.WriteString(m.renderStatusBar())
		return b.String()
	}

	// 数据表格
	b.WriteString(m.renderTable())
	b.WriteString("\n\n")

	// 状态栏
	b.WriteString(m.renderStatusBar())

	return b.String()
}

// renderTitleBar 渲染标题栏
func (m Model) renderTitleBar() string {
	var b strings.Builder

	// 标题
	title := "📈 Stocker 实时监控"

	// 最后更新时间
	updateTime := "未获取"
	if !m.lastUpdate.IsZero() {
		updateTime = m.lastUpdate.Format("2006-01-02 15:04:05")
	}
	timeStr := fmt.Sprintf("最后更新: %s", updateTime)

	// 连接状态
	var statusIcon string
	switch m.connectionState {
	case ConnectionNormal:
		statusIcon = "🟢"
	case ConnectionSlow:
		statusIcon = "🟡"
	case ConnectionFailed:
		statusIcon = "🔴"
	}

	// 组合标题行
	titleLine := fmt.Sprintf("%s %s  %s %s", title, statusIcon, timeStr, m.getStateIndicator())

	if m.simpleMode {
		b.WriteString(titleLine)
	} else {
		b.WriteString(titleStyle.Render(titleLine))
	}

	return b.String()
}

// getStateIndicator 获取状态指示器
func (m Model) getStateIndicator() string {
	if m.paused {
		return "⏸️"
	}
	return ""
}

// renderTable 渲染股票数据表格
func (m Model) renderTable() string {
	var b strings.Builder

	// 根据终端宽度决定显示的列
	showVolume := m.width >= 100
	showAmplitude := m.width >= 120

	// 计算列宽
	colWidth := m.calculateColumnWidths(showVolume, showAmplitude)

	// 表头
	b.WriteString(m.renderTableHeader(colWidth, showVolume, showAmplitude))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", m.getSeparatorWidth(showVolume, showAmplitude)))
	b.WriteString("\n")

	// 数据行
	for i, stock := range m.stockData {
		b.WriteString(m.renderTableRow(i+1, stock, colWidth, showVolume, showAmplitude))
		b.WriteString("\n")
	}

	return b.String()
}

// calculateColumnWidths 计算列宽
func (m Model) calculateColumnWidths(showVolume, showAmplitude bool) ColumnWidths {
	baseWidth := 80
	if m.width > baseWidth {
		baseWidth = m.width
	}

	return ColumnWidths{
		Index:  4,
		Code:   10,
		Name:   12,
		Price:  10,
		Change: 10,
		Volume: 10,
		Amp:    8,
	}
}

// ColumnWidths 列宽配置
type ColumnWidths struct {
	Index  int
	Code   int
	Name   int
	Price  int
	Change int
	Volume int
	Amp    int
}

// renderTableHeader 渲染表头
func (m Model) renderTableHeader(w ColumnWidths, showVolume, showAmplitude bool) string {
	headers := []string{
		padCenter("#", w.Index),
		padCenter("代码", w.Code),
		padCenter("名称", w.Name),
		padCenter("最新价格", w.Price),
		padCenter("涨跌幅", w.Change),
	}

	if showVolume {
		headers = append(headers, padCenter("成交量", w.Volume))
	}

	if showAmplitude {
		headers = append(headers, padCenter("振幅", w.Amp))
	}

	header := strings.Join(headers, " │ ")
	if m.simpleMode {
		return header
	}
	return headerStyle.Render(header)
}

// renderTableRow 渲染数据行
func (m Model) renderTableRow(index int, stock *models.StockData, w ColumnWidths, showVolume, showAmplitude bool) string {
	// 序号
	indexStr := padCenter(fmt.Sprintf("%d", index), w.Index)

	// 代码
	codeStr := padLeft(stock.GetDisplaySymbol(), w.Code)

	// 名称（截断过长的名称）
	nameStr := truncate(stock.Name, w.Name)
	nameStr = padLeft(nameStr, w.Name)

	// 价格
	priceStr := padLeft(models.FormatPrice(stock.Current), w.Price)

	// 涨跌幅
	changeStr := padLeft(stock.FormatChange(), w.Change)

	var styledChange string
	if m.simpleMode {
		// 简洁模式：无颜色
		styledChange = changeStr
	} else {
		// 正常模式：带颜色
		if stock.ChangePercent > 0 {
			styledChange = upStyle.Render(changeStr)
		} else if stock.ChangePercent < 0 {
			styledChange = downStyle.Render(changeStr)
		} else {
			styledChange = neutralStyle.Render(changeStr)
		}
	}

	cells := []string{indexStr, codeStr, nameStr, priceStr, styledChange}

	// 成交量
	if showVolume {
		volumeStr := padLeft(stock.Volume, w.Volume)
		cells = append(cells, volumeStr)
	}

	// 振幅
	if showAmplitude {
		amplitude := ((stock.High - stock.Low) / stock.Close) * 100
		ampStr := padLeft(fmt.Sprintf("%.2f%%", amplitude), w.Amp)
		cells = append(cells, ampStr)
	}

	return strings.Join(cells, " │ ")
}

// renderStatusBar 渲染状态栏
func (m Model) renderStatusBar() string {
	var b strings.Builder

	if m.paused {
		if m.simpleMode {
			b.WriteString("⏸️  监控已暂停")
		} else {
			b.WriteString(warningStyle.Render("⏸️  监控已暂停"))
		}
		b.WriteString(" │ ")
	} else {
		if m.simpleMode {
			b.WriteString("📊 实时刷新中...")
		} else {
			b.WriteString(successStyle.Render("📊 实时刷新中..."))
		}
		b.WriteString(" │ ")
	}

	// 操作提示
	b.WriteString("[空格]暂停 [r]刷新 [s]设置 [h]帮助 [q]退出")

	// 状态消息
	if m.statusMessage != "" {
		b.WriteString("\n")
		if m.simpleMode {
			b.WriteString(m.statusMessage)
		} else {
			if m.pauseReason == PauseSchedule || m.connectionState == ConnectionSlow {
				b.WriteString(warningStyle.Render(m.statusMessage))
			} else if m.err != nil {
				b.WriteString(errorStyle.Render(m.statusMessage))
			} else {
				b.WriteString(successStyle.Render(m.statusMessage))
			}
		}
	}

	return b.String()
}

// getSeparatorWidth 获取分隔线宽度
func (m Model) getSeparatorWidth(showVolume, showAmplitude bool) int {
	w := m.calculateColumnWidths(showVolume, showAmplitude)
	width := w.Index + w.Code + w.Name + w.Price + w.Change + 16 // 16 for separators

	if showVolume {
		width += w.Volume + 3
	}
	if showAmplitude {
		width += w.Amp + 3
	}

	return width
}

// 辅助函数

// padLeft 左对齐填充
func padLeft(s string, width int) string {
	// 计算实际显示宽度（考虑中文字符）
	displayWidth := 0
	for _, r := range s {
		if r > 127 {
			displayWidth += 2 // 中文字符占2个宽度
		} else {
			displayWidth += 1
		}
	}

	if displayWidth >= width {
		return s
	}

	return s + strings.Repeat(" ", width-displayWidth)
}

// padCenter 居中对齐
func padCenter(s string, width int) string {
	// 计算实际显示宽度
	displayWidth := 0
	for _, r := range s {
		if r > 127 {
			displayWidth += 2
		} else {
			displayWidth += 1
		}
	}

	if displayWidth >= width {
		return s
	}

	leftPad := (width - displayWidth) / 2
	rightPad := width - displayWidth - leftPad

	return strings.Repeat(" ", leftPad) + s + strings.Repeat(" ", rightPad)
}

// truncate 截断字符串
func truncate(s string, width int) string {
	displayWidth := 0
	var result strings.Builder

	for _, r := range s {
		charWidth := 1
		if r > 127 {
			charWidth = 2
		}

		if displayWidth+charWidth > width-2 {
			result.WriteString("..")
			break
		}

		result.WriteRune(r)
		displayWidth += charWidth
	}

	if displayWidth < len(s) {
		return result.String()
	}
	return s
}
