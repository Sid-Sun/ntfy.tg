package bot

import (
	"github.com/sid-sun/ntfy.tg/cmd/config"
	"github.com/sid-sun/ntfy.tg/pkg/bot/monitoring"
	"github.com/sid-sun/ntfy.tg/pkg/bot/router"
	"github.com/sid-sun/ntfy.tg/pkg/subscriber"
	subscriptionmanager "github.com/sid-sun/ntfy.tg/pkg/subscription_manager"
	"go.uber.org/zap"
)

// StartBot starts the bot, inits all the requited submodules and routine for shutdown
func StartBot(cfg config.Config, logger *zap.Logger) {
	restartChan := make(chan bool, 1)
	subscriptionmanager.InitSubscriptions(restartChan)
	router := router.New(cfg.Bot, logger)
	botInstance := router.GetBot()
	sub := subscriber.NewSubscriber(botInstance, restartChan, logger)
	go sub.Subscribe()
	go monitoring.PeriodicNotify(logger)

	logger.Info("[StartBot] Started Bot")
	router.NewUpdateChan().ListenAndServe()
}
