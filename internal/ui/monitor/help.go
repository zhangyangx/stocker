package monitor

import (
	"fmt"
	"strings"
)

// renderHelp 渲染帮助界面
func (m Model) renderHelp() string {
	var b strings.Builder

	// 标题
	if m.simpleMode {
		b.WriteString("快捷键帮助 - 简洁模式")
	} else {
		b.WriteString(titleStyle.Render("📖 快捷键帮助"))
	}
	b.WriteString("\n\n")

	// 快捷键列表
	shortcuts := []struct {
		key  string
		desc string
	}{
		{"空格键", "暂停/继续监控"},
		{"r", "手动刷新数据"},
		{"q / ESC", "返回主菜单"},
		{"Ctrl+C", "退出程序"},
	}

	// 非简洁模式添加更多快捷键
	if !m.simpleMode {
		shortcuts = append(shortcuts,
			struct {
				key  string
				desc string
			}{"s", "进入设置界面"},
			struct {
				key  string
				desc string
			}{"h", "显示/隐藏帮助"},
		)
	}

	b.WriteString(strings.Repeat("─", 50))
	b.WriteString("\n")

	for _, sc := range shortcuts {
		b.WriteString(fmt.Sprintf("  %-15s  %s\n", sc.key, sc.desc))
	}

	b.WriteString(strings.Repeat("─", 50))
	b.WriteString("\n\n")

	// 提示
	if m.simpleMode {
		b.WriteString("简洁模式说明\n")
		b.WriteString("  • 页面布局与正常模式一致\n")
		b.WriteString("  • 去掉所有颜色显示\n")
		b.WriteString("  • 按 s 键进入设置可关闭简洁模式\n")
	} else {
		b.WriteString(warningStyle.Render("💡 提示"))
		b.WriteString("\n")
		b.WriteString("  • 监控界面会每隔几秒自动刷新股票数据\n")
		b.WriteString("  • 按空格键可以暂停自动刷新\n")
		b.WriteString("  • 按 r 键可以手动立即刷新\n")
		b.WriteString("  • 按 s 键可以修改刷新间隔等设置（含简洁模式）\n")
	}
	b.WriteString("\n")

	if m.simpleMode {
		b.WriteString("按 h 或 q 键返回监控界面")
	} else {
		b.WriteString(successStyle.Render("按 h 或 q 键返回监控界面"))
	}

	return b.String()
}

// renderSettings 渲染设置界面
func (m Model) renderSettings() string {
	var b strings.Builder

	// 标题
	b.WriteString(titleStyle.Render("⚙️  监控设置"))
	b.WriteString("\n\n")

	b.WriteString(strings.Repeat("─", 60))
	b.WriteString("\n\n")

	// 简洁模式状态
	simpleModeStatus := "关闭"
	if m.config.Preferences.SimpleMode {
		simpleModeStatus = "开启"
	}

	// 设置项
	settings := []struct {
		num   string
		name  string
		value string
		hint  string
	}{
		{
			"1",
			"刷新间隔",
			fmt.Sprintf("%d 秒", m.config.Preferences.RefreshInterval),
			"[+/-] 调整",
		},
		{
			"2",
			"显示列",
			"代码 名称 价格 涨跌幅 成交量",
			"",
		},
		{
			"3",
			"颜色主题",
			"暗黑",
			"",
		},
		{
			"4",
			"排序方式",
			"按添加顺序",
			"",
		},
		{
			"5",
			"简洁模式",
			simpleModeStatus,
			"[空格] 切换",
		},
	}

	for i, setting := range settings {
		indicator := " "
		if i == m.settings.selectedItem {
			indicator = "→"
		}

		b.WriteString(fmt.Sprintf("%s [%s] %-12s: %-30s %s\n",
			indicator,
			setting.num,
			setting.name,
			setting.value,
			setting.hint,
		))
	}

	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 60))
	b.WriteString("\n\n")

	// 操作提示
	b.WriteString("操作说明:\n")
	b.WriteString("  • [↑/↓ 或 k/j] 上下移动选择\n")
	b.WriteString("  • [1-5] 直接选择设置项\n")
	b.WriteString("  • [+/-] 调整刷新间隔\n")
	b.WriteString("  • [空格] 切换简洁模式（去掉颜色）\n")
	b.WriteString("  • [Enter] 保存并返回\n")
	b.WriteString("  • [q/ESC] 取消并返回\n")
	b.WriteString("\n")

	b.WriteString(successStyle.Render("修改会自动保存"))

	return b.String()
}
