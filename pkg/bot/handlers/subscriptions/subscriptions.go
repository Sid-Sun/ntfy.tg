package subscriptions

import (
	"fmt"
	"log/slog"
	"strings"

	subscriptionmanager "github.com/sid-sun/ntfy.tg/pkg/subscription_manager"
	tele "gopkg.in/telebot.v4"
)

// Handler handles all repeat requests
func Handler(c tele.Context) error {
	slog.Info("[Subscriptions] [Attempt]")

	chatID := c.Chat().ID

	// Get chat subscriptions and send them to the user
	topics := subscriptionmanager.GetChatSubscriptions(chatID)
	var msg string
	if len(topics) == 0 {
		msg = "You are not subscribed to any topics, to subscribe to a topic use /subscribe <topic>"
	} else {
		msg = fmt.Sprintf("You are subscribed to: \n- %s", strings.Join(topics, "\n- "))
	}

	if err := c.Reply(msg); err != nil {
		slog.Error(fmt.Sprintf("[Subscriptions] [Send] %s", err.Error()))
		return err
	}

	slog.Info("[Subscriptions] [Success]")
	return nil
}
