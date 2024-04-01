package router

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sid-sun/ntfy.tg/cmd/config"
	"github.com/sid-sun/ntfy.tg/pkg/bot/handlers/subscribe"
	"github.com/sid-sun/ntfy.tg/pkg/bot/handlers/subscriptions"
	"github.com/sid-sun/ntfy.tg/pkg/bot/handlers/unsubscribe"
	"github.com/sid-sun/ntfy.tg/pkg/bot/handlers/unsubscribeall"
	"go.uber.org/zap"
)

type updates struct {
	ch     tgbotapi.UpdatesChannel
	bot    *tgbotapi.BotAPI
	logger *zap.Logger
}

// ListenAndServe starts listens on the update channel and handles routing the update to handlers
func (u updates) ListenAndServe() {
	u.logger.Info(fmt.Sprintf("[StartBot] Started Bot: %s", u.bot.Self.FirstName))
	for update := range u.ch {
		update := update
		go func() {
			if update.Message == nil {
				return
			}
			helpMessage := "Hi! Here are my commands:\n/subscribe <topic> to subscribe to a topic\n/unsubscribe <topic> unsubscribe from a topic\n/unsubscribeall ⚠️ to unsub from all topics ⚠️\n/subscriptions to list your subs"
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "subscribe":
					subscribe.Handler(u.bot, update, u.logger)
					return
				case "unsubscribe":
					unsubscribe.Handler(u.bot, update, u.logger)
				case "unsubscribeall":
					unsubscribeall.Handler(u.bot, update, u.logger)
				case "subscriptions":
					subscriptions.Handler(u.bot, update, u.logger)
				case "start":
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Welcome to ntfy.tg, to subscribe to a topic send: /subscribe <topic> to see help, send: /help")
					u.bot.Send(msg)
				case "help":
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpMessage)
					u.bot.Send(msg)
				default:
				}
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpMessage)
				u.bot.Send(msg)
			}
		}()
	}
}

type bot struct {
	bot    *tgbotapi.BotAPI
	logger *zap.Logger
}

// NewUpdateChan creates a new channel to get update
func (b bot) NewUpdateChan() updates {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	ch := b.bot.GetUpdatesChan(u)
	return updates{ch: ch, bot: b.bot, logger: b.logger}
}

func (b bot) GetBot() *tgbotapi.BotAPI {
	return b.bot
}

// New returns a new instance of the router
func New(cfg config.BotConfig, logger *zap.Logger) bot {
	b, err := tgbotapi.NewBotAPI(cfg.Token())
	if err != nil {
		panic(err)
	}
	return bot{
		bot:    b,
		logger: logger,
	}
}
