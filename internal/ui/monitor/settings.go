package monitor

import (
	"stocker/internal/config"
)

// SettingsModel 设置模型
type SettingsModel struct {
	config       *config.Config
	selectedItem int
	modified     bool
}

// NewSettingsModel 创建新的设置模型
func NewSettingsModel(cfg *config.Config) *SettingsModel {
	return &SettingsModel{
		config:       cfg,
		selectedItem: 0,
		modified:     false,
	}
}
