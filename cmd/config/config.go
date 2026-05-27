package config

import (
	"github.com/spf13/viper"
)

var cfg Config

type StorageEngineConfig struct {
	URL            string
	ObjectID       string
	ObjectPassword string
}

// Config contains all the neccessary configurations
type Config struct {
	StorageEngine StorageEngineConfig
	Bot           BotConfig
	adminChatId   int64
	environment   string
	ntfyDomain    string
	metricsPort   string
}

// GetEnv returns the current developemnt environment
func (c Config) GetEnv() string {
	return c.environment
}

// GetEnv returns the current developemnt environment
func (c Config) GetAdminChatID() int64 {
	return c.adminChatId
}

func (c Config) GetNtfyDomain() string {
	return c.ntfyDomain
}

// GetMetricsPort returns the port for the Prometheus metrics HTTP server
func (c Config) GetMetricsPort() string {
	return c.metricsPort
}

// Load reads all config from env to config
func Load() Config {
	viper.AutomaticEnv()
	viper.SetDefault("NTFY_DOMAIN", "ntfy.sh")
	viper.SetDefault("METRICS_PORT", "9090")
	cfg = Config{
		environment: viper.GetString("APP_ENV"),
		adminChatId: viper.GetInt64("ADMIN_CHAT_ID"),
		Bot: BotConfig{
			tkn: viper.GetString("API_TOKEN"),
		},
		ntfyDomain:  viper.GetString("NTFY_DOMAIN"),
		metricsPort: viper.GetString("METRICS_PORT"),
		StorageEngine: StorageEngineConfig{
			URL:            viper.GetString("SE_URL"),
			ObjectID:       viper.GetString("SE_OBJ_ID"),
			ObjectPassword: viper.GetString("SE_OBJ_PASS"),
		},
	}

	return cfg
}

func GetConfig() Config {
	return cfg
}
