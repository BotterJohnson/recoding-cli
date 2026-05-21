package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Load 按优先级加载配置：默认值 → 全局配置 → 项目配置 → 环境变量 → flag。
func Load() (*Config, error) {
	v := viper.New()

	// 默认值
	v.SetDefault("provider.name", "openai")
	v.SetDefault("provider.base_url", "https://api.openai.com/v1")
	v.SetDefault("provider.model", "gpt-4o")
	v.SetDefault("provider.temperature", 0.7)
	v.SetDefault("provider.max_tokens", 4096)

	// 全局配置: ~/.recoding/config.yaml
	home, err := os.UserHomeDir()
	if err == nil {
		v.AddConfigPath(filepath.Join(home, ".recoding"))
	}

	// 项目配置: .recoding/config.yaml
	v.AddConfigPath(".recoding")

	// 项目 configs 目录
	v.AddConfigPath("configs")

	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// 读取配置文件（先找 config.yaml，找不到再找 default.yaml）
	if err := v.ReadInConfig(); err != nil {
		v.SetConfigName("default")
		_ = v.ReadInConfig()
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config unmarshal: %w", err)
	}

	// 环境变量覆盖（显式处理，避免 Viper 嵌套 key 的 bug）
	if val := os.Getenv("RECODING_API_KEY"); val != "" {
		cfg.Provider.APIKey = val
	}
	if val := os.Getenv("RECODING_BASE_URL"); val != "" {
		cfg.Provider.BaseURL = val
	}
	if val := os.Getenv("RECODING_MODEL"); val != "" {
		cfg.Provider.Model = val
	}

	return &cfg, nil
}
