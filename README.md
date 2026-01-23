# Stocker

一个基于命令行的股票监测工具，提供实时股价监控、配置管理和友好的用户交互界面。

## 功能特性

- **实时监控** - TUI 界面实时显示股票行情，自动刷新
- **股票搜索** - 快速搜索并添加 A 股股票到监控列表
- **多彩显示** - 涨跌用颜色区分，一目了然
- **简洁模式** - 支持无颜色显示，适合不同终端环境
- **可配置** - 自定义刷新间隔、API 提供商等设置
- **跨平台** - 支持 macOS、Windows、Linux

## 快速开始

### 安装

```bash
# 克隆仓库
git clone https://github.com/zhangyangx/stocker.git
cd stocker

# 编译
make mac      # macOS ARM64
make windows  # Windows AMD64
```

编译后的二进制文件位于 `build/` 目录。

### 运行

```bash
# 直接运行
./build/stocker

# 或使用 go run
go run ./cmd
```

## 使用方法

### 命令行选项

```bash
stocker [选项]

选项:
  -h, -help     显示帮助信息
  -v, -version  显示版本信息
```

### 交互式命令

启动后进入交互式界面，支持以下命令：

| 命令 | 说明 |
|------|------|
| `help`, `h` | 显示帮助信息 |
| `monitor`, `m` / `start` | 启动实时监控界面 |
| `list`, `ls` | 显示监控列表 |
| `list -a`, `ls -a` | 显示详细信息（含实时行情） |
| `remove`, `rm <序号>` | 按序号删除股票 |
| `search`, `s <关键词>` | 搜索并添加股票到监控列表 |
| `refresh`, `r` | 刷新股票实时数据 |
| `clear` | 清空所有监控股票 |
| `config` | 查看或修改系统配置 |
| `exit`, `q` | 退出程序 |

### 监控界面快捷键

在监控界面中，可以使用以下快捷键：

| 按键 | 说明 |
|------|------|
| `空格` | 暂停/继续监控 |
| `r` | 手动刷新数据 |
| `s` | 进入设置界面 |
| `h` | 显示/隐藏帮助 |
| `q` / `ESC` | 返回主菜单 |
| `Ctrl+C` | 退出程序 |

### 设置界面

在设置界面中可以配置：

- **刷新间隔** - 使用 `+`/`-` 调整自动刷新时间
- **简洁模式** - 按空格键切换（去除颜色显示）

配置会自动保存到 `~/.config/stocker/config.yaml`

## 配置文件

配置文件位于 `~/.config/stocker/config.yaml`，首次运行会自动创建。

默认配置示例：

```yaml
app:
  name: Stocker
  version: 1.0.0-beta

preferences:
  refresh_interval: 3   # 刷新间隔（秒）
  simple_mode: false    # 简洁模式（无颜色）

api:
  timeout: 5s
  retry_count: 3
  retry_delay: 1s
  max_concurrent: 10
  cache_duration: 2s
  primary_provider: sina
  fallback_provider: ""

stocks:
  watchlist: []  # 监控的股票列表
```

## 数据来源

目前支持以下数据提供商：

- **新浪财经** (sina) - 默认提供商
- **腾讯财经** (tencent) - 备选提供商

## 项目结构

```
stocker/
├── cmd/              # 程序入口
├── internal/
│   ├── app/         # 应用程序主逻辑
│   ├── config/      # 配置管理
│   ├── data/        # 数据获取和解析
│   ├── market/      # 市场相关逻辑
│   ├── stock/       # 股票管理
│   └── ui/          # 用户界面
│       ├── cli.go        # 命令行界面
│       ├── commands.go   # 命令定义
│       └── monitor/      # TUI 监控界面
├── pkg/             # 公共包
│   └── models/      # 数据模型
├── docs/            # 文档
├── Makefile         # 构建脚本
└── go.mod           # Go 模块定义
```

## 开发

```bash
# 安装依赖
go mod download

# 运行测试
go test ./...

# 格式化代码
go fmt ./...

# 本地运行
go run ./cmd
```

## 系统要求

- Go 1.21 或更高版本
- 支持的操作系统：macOS、Linux、Windows
