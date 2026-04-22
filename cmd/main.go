package main

import (
	"github.com/sid-sun/ntfy.tg/cmd/config"
	"github.com/sid-sun/ntfy.tg/pkg/bot"
)

func main() {
	cfg := config.Load()
	initLogger(cfg.GetEnv())
	bot.StartBot(cfg, logger)
}
