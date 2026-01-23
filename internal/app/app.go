package app

import (
	"fmt"

	"stocker/internal/config"
	"stocker/internal/ui"
)

// Application 应用程序主结构体
type Application struct {
	config  *config.Config
	cli     *ui.CLI
	version string
}

// NewApplication 创建新的应用程序实例
func NewApplication(version string) (*Application, error) {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	// 统一版本来源（优先使用构建传入版本）
	if version != "" {
		cfg.App.Version = version
	}

	// 创建CLI实例
	cli, err := ui.NewCLI(cfg, version)
	if err != nil {
		return nil, fmt.Errorf("创建CLI失败: %w", err)
	}

	return &Application{
		config:  cfg,
		cli:     cli,
		version: version,
	}, nil
}

// Run 运行应用程序
func (app *Application) Run() error {
	// 启动CLI界面
	return app.cli.Start()
}

// Shutdown 关闭应用程序
func (app *Application) Shutdown() error {
	if app.cli != nil {
		app.cli.Stop()
	}

	// 保存配置
	if app.config != nil {
		if err := app.config.Save(); err != nil {
			return fmt.Errorf("保存配置失败: %w", err)
		}
	}

	return nil
}

// GetConfig 获取当前配置
func (app *Application) GetConfig() *config.Config {
	return app.config
}

// GetVersion 获取应用程序版本
func (app *Application) GetVersion() string {
	return app.version
}
