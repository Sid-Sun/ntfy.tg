package config

import (
	"strings"

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
	PingURLs      []string
}

// GetEnv returns the current developemnt environment
func (c Config) GetEnv() string {
	return c.environment
}

// GetEnv returns the current developemnt environment
func (c Config) GetAdminChatID() int64 {
	return c.adminChatId
}

// Load reads all config from env to config
func Load() Config {
	viper.AutomaticEnv()
	cfg = Config{
		environment: viper.GetString("APP_ENV"),
		adminChatId: viper.GetInt64("ADMIN_CHAT_ID"),
		Bot: BotConfig{
			tkn: viper.GetString("API_TOKEN"),
		},
		StorageEngine: StorageEngineConfig{
			URL:            viper.GetString("SE_URL"),
			ObjectID:       viper.GetString("SE_OBJ_ID"),
			ObjectPassword: viper.GetString("SE_OBJ_PASS"),
		},
		PingURLs: strings.Split(viper.GetString("PING_URLS"), ";"),
	}

	return cfg
}

func GetConfig() Config {
	return cfg
}
