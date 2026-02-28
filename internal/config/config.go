package config

import "time"

// Config holds all application configuration.
type Config struct {
	Completion ProviderConfig `toml:"completion"`
	Embedding  ProviderConfig `toml:"embedding"`
	Storage    StorageConfig  `toml:"storage"`
	Server     ServerConfig   `toml:"server"`
}

// ProviderConfig configures an LLM provider.
type ProviderConfig struct {
	Provider string `toml:"provider"` // "openai", "anthropic", "ollama"
	APIKey   string `toml:"api_key"`
	Model    string `toml:"model"`
	BaseURL  string `toml:"base_url"` // override for Ollama or proxies
}

// StorageConfig configures the SQLite database.
type StorageConfig struct {
	DBPath string `toml:"db_path"`
}

// ServerConfig configures the web server.
type ServerConfig struct {
	Port    int           `toml:"port"`
	Timeout time.Duration `toml:"timeout"`
}

// Defaults returns a Config with sensible default values.
func Defaults() Config {
	return Config{
		Completion: ProviderConfig{
			Provider: "openai",
		},
		Embedding: ProviderConfig{
			Provider: "openai",
		},
		Storage: StorageConfig{
			DBPath: "healthanalyzer.db",
		},
		Server: ServerConfig{
			Port:    8080,
			Timeout: 30 * time.Second,
		},
	}
}
