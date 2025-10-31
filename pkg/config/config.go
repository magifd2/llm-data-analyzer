package config

import (
	"github.com/spf13/viper"
)

// EndpointConfig defines the configuration for a single LLM endpoint.
type EndpointConfig struct {
	Name              string `mapstructure:"name"`
	EndpointURL       string `mapstructure:"endpoint_url"`
	APIKeyEnv         string `mapstructure:"api_key_env"`
	Model             string `mapstructure:"model"`
	ContextWindowSize int    `mapstructure:"context_window_size"`
	ChunkSize         int    `mapstructure:"chunk_size"`
}

// Config defines the overall configuration for the application.
type Config struct {
	Endpoints []EndpointConfig `mapstructure:"endpoints"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config Config, err error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	if err = viper.ReadInConfig(); err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
