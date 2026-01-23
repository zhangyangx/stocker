package ui

import (
	"fmt"
	"strings"
)

// CommandType 命令类型枚举
type CommandType int

const (
	CommandHelp CommandType = iota
	CommandExit
	CommandConfig
	CommandRemove
	CommandList
	CommandSearch
	CommandWatch
	CommandStop
	CommandClear
	CommandRefresh
	CommandMonitor
	CommandStart
	CommandUnknown
)

// Command 命令结构
type Command struct {
	Type CommandType
	Args []string
}

// ParseCommand 解析用户输入的命令
func ParseCommand(input string) *Command {
	input = strings.TrimSpace(input)
	if input == "" {
		return &Command{Type: CommandUnknown}
	}

	parts := strings.Fields(input)
	if len(parts) == 0 {
		return &Command{Type: CommandUnknown}
	}

	command := strings.ToLower(parts[0])
	var args []string
	if len(parts) > 1 {
		args = parts[1:]
	}

	switch command {
	case "help", "h":
		return &Command{Type: CommandHelp, Args: args}
	case "exit", "q":
		return &Command{Type: CommandExit, Args: args}
	case "config":
		return &Command{Type: CommandConfig, Args: args}
	case "remove", "rm":
		return &Command{Type: CommandRemove, Args: args}
	case "list", "ls":
		return &Command{Type: CommandList, Args: args}
	case "search", "s":
		return &Command{Type: CommandSearch, Args: args}
	case "watch", "w":
		return &Command{Type: CommandWatch, Args: args}
	case "stop":
		return &Command{Type: CommandStop, Args: args}
	case "clear":
		return &Command{Type: CommandClear, Args: args}
	case "refresh", "r":
		return &Command{Type: CommandRefresh, Args: args}
	case "monitor", "m":
		return &Command{Type: CommandMonitor, Args: args}
	case "start":
		return &Command{Type: CommandStart, Args: args}
	default:
		return &Command{Type: CommandUnknown, Args: args}
	}
}

// GetCommandHelp 获取命令帮助信息
func GetCommandHelp() string {
	const (
		colCmd  = 20
		colDesc = 40
	)

	var builder strings.Builder

	// 标题
	builder.WriteString("\n" + strings.Repeat("═", 65) + "\n")
	builder.WriteString("                       📖 命令帮助\n")
	builder.WriteString(strings.Repeat("═", 65) + "\n\n")

	// 命令列表
	commands := []struct {
		cmd  string
		desc string
	}{
		{"help, h", "显示此帮助信息"},
		{"monitor, m / start", "启动实时监控界面"},
		{"list, ls", "显示监控列表并支持删除操作"},
		{"list -a, ls -a", "显示详细信息（含实时行情）并支持删除"},
		{"remove, rm <序号>", "按序号删除股票"},
		{"search, s <关键词>", "搜索并添加股票到监控列表"},
		{"refresh, r", "刷新股票实时数据"},
		{"clear", "清空所有监控股票"},
		{"config", "查看或修改系统配置"},
		{"exit, q", "退出程序"},
	}

	for _, cmd := range commands {
		builder.WriteString("  " + padRight(cmd.cmd, colCmd) + "  " + cmd.desc + "\n")
	}

	builder.WriteString("\n" + strings.Repeat("─", 65) + "\n")
	builder.WriteString("💡 提示: 在列表页面可以输入序号快速操作股票\n")
	builder.WriteString("💡 提示: 在监控界面按 's' 键进入设置，可切换简洁模式（去掉颜色）\n")
	builder.WriteString(strings.Repeat("═", 65) + "\n\n")

	return builder.String()
}

// FormatSuccessMessage 格式化成功消息
func FormatSuccessMessage(message string) string {
	return fmt.Sprintf("✓ %s", message)
}

// FormatErrorMessage 格式化错误消息
func FormatErrorMessage(message string) string {
	return fmt.Sprintf("✗ %s", message)
}

// FormatWarningMessage 格式化警告消息
func FormatWarningMessage(message string) string {
	return fmt.Sprintf("⚠ %s", message)
}

// FormatInfoMessage 格式化信息消息
func FormatInfoMessage(message string) string {
	return fmt.Sprintf("ℹ %s", message)
}

// GetExitMessage 获取统一的退出消息
func GetExitMessage() string {
	width := 50
	message := "感谢使用 Stocker"

	var builder strings.Builder
	builder.WriteString("\n" + strings.Repeat("─", width) + "\n")

	messagePadding := (width - displayWidth(message) - 4) / 2
	builder.WriteString(strings.Repeat(" ", messagePadding))
	builder.WriteString("👋 " + message + "\n")

	builder.WriteString(strings.Repeat("─", width) + "\n\n")

	return builder.String()
}
