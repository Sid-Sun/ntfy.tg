package main

import (
	"github.com/sid-sun/ntfy.tg/cmd/config"
	"github.com/sid-sun/ntfy.tg/pkg/bot"
	"go.uber.org/zap"
)

var logger *zap.Logger

func main() {
	cfg := config.Load()
	initLogger(cfg.GetEnv())
	bot.StartBot(cfg, logger)
}

func initLogger(env string) {
	var err error

	if env == "dev" {
		logger, err = zap.NewDevelopmentConfig().Build()
	} else {
		logger, err = zap.NewProductionConfig().Build()
	}

	if err != nil {
		panic(err)
	}
}
