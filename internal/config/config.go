package config

// Config 是应用的顶层配置结构。
type Config struct {
	Provider ProviderConfig `mapstructure:"provider"`
}

// ProviderConfig 是 LLM Provider 的配置。
type ProviderConfig struct {
	Name        string  `mapstructure:"name"`
	APIKey      string  `mapstructure:"api_key"`
	BaseURL     string  `mapstructure:"base_url"`
	Model       string  `mapstructure:"model"`
	Temperature float64 `mapstructure:"temperature"`
	MaxTokens   int     `mapstructure:"max_tokens"`
}
