package ui

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"stocker/internal/config"
	"stocker/internal/data"
	"stocker/internal/stock"
	"stocker/internal/ui/monitor"
	"stocker/pkg/models"
)

// CLI 命令行界面结构体
type CLI struct {
	config    *config.Config
	reader    *bufio.Reader
	isRunning bool
	version   string
	manager   *stock.Manager
}

// NewCLI 创建新的CLI实例
func NewCLI(cfg *config.Config, version string) (*CLI, error) {
	// 创建股票管理器
	manager, err := stock.NewManager(cfg)
	if err != nil {
		return nil, fmt.Errorf("创建股票管理器失败: %w", err)
	}

	return &CLI{
		config:    cfg,
		reader:    bufio.NewReader(os.Stdin),
		isRunning: false,
		version:   version,
		manager:   manager,
	}, nil
}

// Start 启动CLI界面
func (cli *CLI) Start() error {
	cli.isRunning = true
	cli.showWelcome()

	// 设置信号处理
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n" + GetExitMessage())
		cli.isRunning = false
		cancel()
		// 给一点时间让消息显示
		time.Sleep(100 * time.Millisecond)
		os.Exit(0)
	}()

	// 输入通道
	inputChan := make(chan string)
	errChan := make(chan error)

	// 命令循环
	for cli.isRunning {
		// 在 goroutine 中读取输入
		go func() {
			input, err := cli.readInput()
			if err != nil {
				errChan <- err
				return
			}
			inputChan <- input
		}()

		// 等待输入或取消
		select {
		case <-ctx.Done():
			return nil
		case err := <-errChan:
			if cli.isRunning {
				fmt.Println(FormatErrorMessage("读取输入失败: " + err.Error()))
			}
		case input := <-inputChan:
			cli.handleInput(input)
		}
	}

	// 清理资源
	cli.manager.Close()
	return nil
}

// Stop 停止CLI界面
func (cli *CLI) Stop() {
	if cli.isRunning {
		fmt.Println(GetExitMessage())
		cli.isRunning = false
	}
	cli.manager.Close()
}

// showWelcome 显示欢迎界面
func (cli *CLI) showWelcome() {
	const width = 60

	fmt.Println()
	fmt.Println(strings.Repeat("═", width))

	// 标题
	title := fmt.Sprintf("📈 Stocker v%s", cli.config.App.Version)
	fmt.Println(centerText(title, width))

	// 副标题
	subtitle := "实时股票监控工具"
	fmt.Println(centerText(subtitle, width))

	fmt.Println(strings.Repeat("═", width))

	// 统计信息
	stats := cli.manager.GetStats()
	var stockInfo string
	if stats.TotalStocks == 0 {
		stockInfo = "当前监控: 暂无股票"
	} else {
		stockInfo = fmt.Sprintf("当前监控: %d 只股票", stats.TotalStocks)
	}
	fmt.Println(centerText(stockInfo, width))

	fmt.Println(strings.Repeat("─", width))

	// 快速提示
	fmt.Println("\n💡 快速开始:")
	fmt.Println("  • 输入 'help' 或 'h' 查看所有命令")
	fmt.Println("  • 输入 'monitor' 或 'start' 启动实时监控")
	fmt.Println("  • 输入 'search 茅台' 搜索并添加股票")
	fmt.Println("  • 输入 'ls -a' 查看实时行情")
	fmt.Println("  • 按 Ctrl+C 或输入 'exit' 退出")

	fmt.Println(strings.Repeat("═", width))
	fmt.Println()
}

// centerText 居中文本（使用空格填充）
func centerText(text string, width int) string {
	textWidth := displayWidth(text)
	if textWidth >= width {
		return text
	}
	leftPad := (width - textWidth) / 2
	return strings.Repeat(" ", leftPad) + text
}

// showPrompt 显示命令提示符
func (cli *CLI) showPrompt() {
	fmt.Print("stocker> ")
}

// readInput 读取用户输入
func (cli *CLI) readInput() (string, error) {
	cli.showPrompt()

	// 使用bufio读取输入
	input, err := cli.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

// handleInput 处理用户输入
func (cli *CLI) handleInput(input string) {
	if strings.TrimSpace(input) == "" {
		return
	}

	command := ParseCommand(input)
	cli.executeCommand(command)
}

// executeCommand 执行命令
func (cli *CLI) executeCommand(cmd *Command) {
	switch cmd.Type {
	case CommandHelp:
		fmt.Print(GetCommandHelp())

	case CommandExit:
		cli.Stop()

	case CommandConfig:
		cli.handleConfigCommand(cmd.Args)

	case CommandMonitor, CommandStart:
		cli.handleMonitorCommand(cmd.Args)

	case CommandRemove:
		cli.handleRemoveCommand(cmd.Args)

	case CommandList:
		cli.handleListCommand(cmd.Args)

	case CommandSearch:
		cli.handleSearchCommand(cmd.Args)

	case CommandClear:
		cli.handleClearCommand(cmd.Args)

	case CommandRefresh:
		cli.handleRefreshCommand(cmd.Args)

	case CommandUnknown:
		fmt.Println(FormatErrorMessage("未知命令。输入 'help' 查看可用命令。"))
	}
}

// handleConfigCommand 处理配置相关命令
func (cli *CLI) handleConfigCommand(args []string) {
	if len(args) == 0 {
		// 显示当前配置
		cli.showConfig()
		return
	}

	switch args[0] {
	case "refresh":
		if len(args) < 2 {
			fmt.Println(FormatErrorMessage("请指定刷新间隔（秒）"))
			fmt.Println("用法: config refresh <秒>")
			return
		}
		cli.setRefreshInterval(args[1])

	default:
		fmt.Println(FormatErrorMessage("未知的配置命令"))
		fmt.Println("可用命令: refresh")
	}
}

// handleRemoveCommand 处理删除股票命令
func (cli *CLI) handleRemoveCommand(args []string) {
	if len(args) == 0 {
		fmt.Println(FormatErrorMessage("请指定序号"))
		fmt.Println("用法: remove <序号>")
		fmt.Println("提示: 也可以使用 'ls' 或 'ls -a' 命令查看列表后选择删除")
		return
	}

	keyword := args[0]

	// 尝试解析为序号
	if index, err := strconv.Atoi(keyword); err == nil {
		// 按序号删除
		cli.removeStockByIndex(index - 1) // 显示序号从1开始
		return
	}

	fmt.Println(FormatErrorMessage("无效的序号"))
}

// handleListCommand 处理显示列表命令
func (cli *CLI) handleListCommand(args []string) {
	showAll := false
	for _, a := range args {
		if a == "-a" || a == "--all" {
			showAll = true
			break
		}
	}

	if showAll {
		cli.displayStockListDetailed()
		cli.promptForAction("detailed")
	} else {
		cli.displayWatchlistLocal()
		cli.promptForAction("local")
	}
}

// handleSearchCommand 处理搜索股票命令
func (cli *CLI) handleSearchCommand(args []string) {
	if len(args) == 0 {
		fmt.Println(FormatErrorMessage("请指定搜索关键词"))
		fmt.Println("用法: search <关键词>")
		return
	}

	keyword := strings.Join(args, " ")
	cli.searchAndAddStock(keyword)
}

// handleClearCommand 处理清空列表命令
func (cli *CLI) handleClearCommand(args []string) {
	cli.clearWatchlist()
}

// handleRefreshCommand 处理刷新数据命令
func (cli *CLI) handleRefreshCommand(args []string) {
	cli.refreshStockData()
}

// handleMonitorCommand 处理实时监控命令
func (cli *CLI) handleMonitorCommand(args []string) {
	// 检查监控列表
	watchlist := cli.manager.GetWatchlist()
	if len(watchlist) == 0 {
		fmt.Println(FormatWarningMessage("监控列表为空，请先添加股票"))
		fmt.Println("提示: 使用 'search <关键词>' 搜索并添加股票")
		return
	}

	// 启动实时监控（从配置读取简洁模式设置）
	simpleMode := cli.config.Preferences.SimpleMode
	if simpleMode {
		fmt.Println("正在启动监控界面（简洁模式）...")
	} else {
		fmt.Println("正在启动监控界面...")
	}

	err := monitor.Start(cli.manager, cli.config, simpleMode)
	if err != nil {
		fmt.Println(FormatErrorMessage(fmt.Sprintf("启动监控失败: %v", err)))
		return
	}

	// 监控退出后，重新显示欢迎界面
	fmt.Println("\n已退出监控模式")
}

// searchAndAddStock 搜索并添加股票
func (cli *CLI) searchAndAddStock(keyword string) {
	fmt.Printf("正在搜索 '%s' 相关股票...\n", keyword)

	results, err := cli.manager.SearchStock(keyword)
	if err != nil {
		fmt.Println(FormatErrorMessage(fmt.Sprintf("搜索失败: %v", err)))
		return
	}

	if len(results) == 0 {
		fmt.Println(FormatWarningMessage(fmt.Sprintf("未找到包含 '%s' 的股票", keyword)))
		return
	}

	// 显示搜索结果
	fmt.Print(DisplaySearchResults(results, keyword))

	// 友好的操作提示
	fmt.Println("\n" + strings.Repeat("─", 50))
	fmt.Println("🔍 搜索结果操作:")
	fmt.Printf("  • 输入序号 (1-%d) 可添加对应股票到监控列表\n", len(results))
	fmt.Println("  • 输入 0 可取消操作")
	fmt.Println("  • 直接按回车可返回主菜单")
	fmt.Print(strings.Repeat("─", 50) + "\n请输入操作: ")

	input, err := cli.readInput()
	if err != nil {
		fmt.Println(FormatErrorMessage("读取输入失败"))
		return
	}

	// 如果用户直接按回车，返回主菜单
	if strings.TrimSpace(input) == "" {
		fmt.Println("已返回主菜单")
		return
	}

	index, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println(FormatErrorMessage("无效的输入，请输入序号数字"))
		return
	}

	if index < 0 || index > len(results) {
		fmt.Println(FormatErrorMessage("序号超出范围"))
		return
	}

	if index == 0 {
		fmt.Println("操作已取消")
		return
	}

	selectedStock := results[index-1]
	err = cli.manager.AddStock(selectedStock.Code)
	if err != nil {
		fmt.Println(FormatErrorMessage(err.Error()))
		return
	}

	fmt.Println(FormatSuccessMessage(fmt.Sprintf("已添加: %s - %s", data.FormatStockCode(selectedStock.Code), selectedStock.Name)))
}

// removeStockByIndex 按序号删除股票
func (cli *CLI) removeStockByIndex(index int) {
	watchlist := cli.manager.GetWatchlist()
	if index < 0 || index >= len(watchlist) {
		fmt.Println(FormatErrorMessage("无效的序号"))
		return
	}

	stock := watchlist[index]
	fmt.Printf("确认删除 %s - %s? (y/N): ", stock.Symbol, stock.Name)

	input, err := cli.readInput()
	if err != nil {
		fmt.Println(FormatErrorMessage("读取输入失败"))
		return
	}

	if strings.ToLower(input) != "y" && strings.ToLower(input) != "yes" {
		fmt.Println("操作已取消")
		return
	}

	err = cli.manager.RemoveStockByIndex(index)
	if err != nil {
		fmt.Println(FormatErrorMessage(err.Error()))
		return
	}

	fmt.Println(FormatSuccessMessage(fmt.Sprintf("已删除: %s - %s", stock.Symbol, stock.Name)))
}

// displayWatchlistLocal 显示股票列表
func (cli *CLI) displayWatchlistLocal() {
	watchlist := cli.manager.GetWatchlist()
	if len(watchlist) == 0 {
		fmt.Println("当前监控股票列表为空")
		return
	}

	fmt.Print(DisplayWatchlistLocal(watchlist, DefaultTableConfig()))
	fmt.Println() // 确保换行
}

// promptForAction 提示用户进行操作
func (cli *CLI) promptForAction(listType string) {
	var watchlist interface{}
	switch listType {
	case "local":
		watchlist = cli.manager.GetSortedWatchlist()
	case "detailed":
		stockData, err := cli.manager.GetAllStockData()
		if err != nil || len(stockData) == 0 {
			return
		}
		watchlist = stockData
	}

	// 检查列表是否为空
	items := 0
	switch v := watchlist.(type) {
	case []models.Stock:
		items = len(v)
	case []*models.StockData:
		items = len(v)
	}

	if items == 0 {
		return
	}

	// 友好的操作提示
	fmt.Println("\n" + strings.Repeat("─", 50))
	fmt.Println("📋 操作提示:")
	fmt.Printf("  • 输入序号 (1-%d) 可删除对应股票\n", items)
	fmt.Println("  • 输入 0 可取消操作")
	fmt.Println("  • 直接按回车可返回主菜单")
	fmt.Print(strings.Repeat("─", 50) + "\n请输入操作: ")

	input, err := cli.readInput()
	if err != nil {
		fmt.Println(FormatErrorMessage("读取输入失败"))
		return
	}

	// 如果用户直接按回车，返回主菜单
	if strings.TrimSpace(input) == "" {
		fmt.Println("已返回主菜单")
		return
	}

	index, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println(FormatErrorMessage("无效的输入，请输入序号数字"))
		return
	}

	if index < 0 || index > items {
		fmt.Println(FormatErrorMessage("序号超出范围"))
		return
	}

	if index == 0 {
		fmt.Println("操作已取消")
		return
	}

	cli.removeStockByIndex(index - 1)
}

// displayStockListDetailed 显示股票列表（包含实时行情）
func (cli *CLI) displayStockListDetailed() {
	// 获取股票数据
	stockData, err := cli.manager.GetAllStockData()
	if err != nil {
		fmt.Println(FormatErrorMessage(fmt.Sprintf("获取股票数据失败: %v", err)))
		return
	}

	if len(stockData) == 0 {
		fmt.Println("当前监控股票列表为空")
		return
	}

	// 显示表格
	fmt.Print(DisplayStockList(stockData, DefaultTableConfig()))
	fmt.Println() // 确保换行
}

// clearWatchlist 清空监控列表
func (cli *CLI) clearWatchlist() {
	watchlist := cli.manager.GetWatchlist()
	if len(watchlist) == 0 {
		fmt.Println(FormatWarningMessage("监控列表已经为空"))
		return
	}

	fmt.Printf("确认清空所有 %d 只股票? (y/N): ", len(watchlist))
	input, err := cli.readInput()
	if err != nil {
		fmt.Println(FormatErrorMessage("读取输入失败"))
		return
	}

	if strings.ToLower(input) != "y" && strings.ToLower(input) != "yes" {
		fmt.Println("操作已取消")
		return
	}

	// 逐个删除
	for _, stock := range watchlist {
		cli.manager.RemoveStock(stock.Symbol)
	}

	fmt.Println(FormatSuccessMessage("已清空监控列表"))
}

// refreshStockData 刷新股票数据
func (cli *CLI) refreshStockData() {
	fmt.Println("正在刷新股票数据...")

	// 清空缓存
	cli.manager.ClearCache()

	// 重新获取数据
	cli.displayStockListDetailed()

	fmt.Println(FormatSuccessMessage("数据刷新完成"))
}

// showConfig 显示当前配置
func (cli *CLI) showConfig() {
	fmt.Println("当前配置:")
	fmt.Printf("  应用名称: %s\n", cli.config.App.Name)
	fmt.Printf("  版本: %s\n", cli.config.App.Version)
	fmt.Printf("  刷新间隔: %d 秒\n", cli.config.Preferences.RefreshInterval)

	stats := cli.manager.GetStats()
	if stats.TotalStocks == 0 {
		fmt.Println("  监控股票: 无")
	} else {
		fmt.Printf("  监控股票: %d 只股票\n", stats.TotalStocks)
		cacheStats := cli.manager.GetCacheStats()
		fmt.Printf("  缓存统计: %s\n", cacheStats.String())
	}
}

// setRefreshInterval 设置刷新间隔
func (cli *CLI) setRefreshInterval(value string) {
	var seconds int
	if _, err := fmt.Sscanf(value, "%d", &seconds); err != nil || seconds <= 0 {
		fmt.Println(FormatErrorMessage("刷新间隔必须是正整数"))
		return
	}

	cli.config.Preferences.RefreshInterval = seconds
	if err := cli.config.Save(); err != nil {
		fmt.Println(FormatErrorMessage("保存配置失败: " + err.Error()))
		return
	}

	fmt.Println(FormatSuccessMessage(fmt.Sprintf("刷新间隔已设置为 %d 秒", seconds)))
}

// 辅助函数
