package router

import (
	"fmt"
	"time"

	"github.com/sid-sun/ntfy.tg/cmd/config"
	"github.com/sid-sun/ntfy.tg/pkg/bot/handlers/subscribe"
	"github.com/sid-sun/ntfy.tg/pkg/bot/handlers/subscriptions"
	"github.com/sid-sun/ntfy.tg/pkg/bot/handlers/unsubscribe"
	"github.com/sid-sun/ntfy.tg/pkg/bot/handlers/unsubscribeall"
	"github.com/sid-sun/ntfy.tg/pkg/metrics"
	"go.uber.org/zap"
	tele "gopkg.in/telebot.v4"
)

type bot struct {
	bot    *tele.Bot
	logger *zap.Logger
}

// ListenAndServe starts listens on the update channel and handles routing the update to handlers
func (b bot) ListenAndServe() {
	b.logger.Info(fmt.Sprintf("[StartBot] Started Bot: %s", b.bot.Me.Username))
	// niceties
	b.bot.Handle("/start", func(ctx tele.Context) error {
		metrics.RecordBotCommand("start")
		return ctx.Reply("Welcome to ntfy.tg, to subscribe to a topic send: /subscribe <topic> to see help, send: /help")
	})
	helpMessage := "Hi! Here are my commands:\n/subscribe <topic> to subscribe to a topic\n/unsubscribe <topic> unsubscribe from a topic\n/unsubscribeall ⚠️ to unsub from all topics ⚠️\n/subscriptions to list your subs"
	helpHandler := func(ctx tele.Context) error {
		metrics.RecordBotCommand("help")
		return ctx.Reply(helpMessage)
	}
	b.bot.Handle("/help", helpHandler)
	b.bot.Handle(tele.OnText, helpHandler)
	// actual handlers
	b.bot.Handle("/subscribe", subscribe.Hnadler)
	b.bot.Handle("/unsubscribe", unsubscribe.Handler)
	b.bot.Handle("/subscriptions", subscriptions.Handler)
	b.bot.Handle("/unsubscribeall", unsubscribeall.Handler)
	b.bot.Start()
}

func (b bot) GetBot() *tele.Bot {
	return b.bot
}

// New returns a new instance of the router
func New(cfg config.BotConfig, logger *zap.Logger) bot {
	pref := tele.Settings{
		Token:  cfg.Token(),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		panic(err)
	}

	return bot{
		bot:    b,
		logger: logger,
	}
}
