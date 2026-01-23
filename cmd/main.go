package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"flag"
	"stocker/internal/app"
)

const (
	AppName    = "Stocker"
	AppVersion = "1.0.0-beta"
)

// BuildTime 由构建系统注入（例如: 2026-02-07 12:00:00）
var BuildTime = ""

func main() {
	// 解析命令行参数
	var showHelp bool
	var showVersion bool

	flag.BoolVar(&showHelp, "help", false, "显示帮助信息")
	flag.BoolVar(&showHelp, "h", false, "显示帮助信息")
	flag.BoolVar(&showVersion, "version", false, "显示版本信息")
	flag.BoolVar(&showVersion, "v", false, "显示版本信息")
	flag.Parse()

	if showHelp {
		printHelp()
		return
	}

	if showVersion {
		printVersion()
		return
	}

	// 创建应用程序实例
	application, err := app.NewApplication(AppVersion)
	if err != nil {
		log.Fatalf("创建应用程序失败: %v", err)
	}

	// 启动应用程序（阻塞直到用户退出）
	if err := application.Run(); err != nil {
		log.Printf("应用程序运行错误: %v", err)
		os.Exit(1)
	}

	// 正常退出
	os.Exit(0)
}

// printHelp 显示帮助信息
func printHelp() {
	fmt.Printf(`%s - 股票监测工具

用法:
  stocker [选项]

选项:
  -h, -help     显示此帮助信息
  -v, -version  显示版本信息

描述:
  Stocker 是一个基于命令行的股票监测工具，提供实时股价监控、
  配置管理和友好的用户交互界面。

使用示例:
  stocker              # 启动交互式界面
  stocker -h           # 显示帮助信息
  stocker -v           # 显示版本信息

更多信息请访问项目主页。
`, AppName)
}

// printVersion 显示版本信息
func printVersion() {
	fmt.Printf("%s %s\n", AppName, AppVersion)
	fmt.Printf("Go版本: %s\n", runtime.Version())
	if BuildTime == "" {
		fmt.Printf("构建时间: unknown\n")
		return
	}
	fmt.Printf("构建时间: %s\n", BuildTime)
}
