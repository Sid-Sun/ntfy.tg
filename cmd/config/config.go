package config

import (
	"os"
	"strconv"
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
	adminChatId, _ := strconv.ParseInt(os.Getenv("ADMIN_CHAT_ID"), 10, 64)
	cfg = Config{
		environment: os.Getenv("APP_ENV"),
		adminChatId: adminChatId,
		Bot: BotConfig{
			tkn: os.Getenv("API_TOKEN"),
		},
		StorageEngine: StorageEngineConfig{
			URL:            os.Getenv("SE_URL"),
			ObjectID:       os.Getenv("SE_OBJ_ID"),
			ObjectPassword: os.Getenv("SE_OBJ_PASS"),
		},
	}
	return cfg
}

func GetConfig() Config {
	return cfg
}
