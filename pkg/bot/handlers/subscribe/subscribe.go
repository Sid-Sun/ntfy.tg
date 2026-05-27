package subscribe

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/sid-sun/ntfy.tg/pkg/metrics"
	subscriptionmanager "github.com/sid-sun/ntfy.tg/pkg/subscription_manager"
	tele "gopkg.in/telebot.v4"
)

// Handler handles all repeat requests
func Hnadler(c tele.Context) error {
	metrics.RecordBotCommand("subscribe")
	slog.Info("[Subscribe] [Attempt]")

	chatID := c.Chat().ID
	topic := c.Message().Payload
	if topic == "" {
		c.Reply("invalid message, to subscribe send: /subscribe <topic>")
		return errors.New("no topic provided")
	}

	validTopic := allowedTopicRegex.MatchString(topic)
	if !validTopic {
		c.Reply("invalid topic, to subscribe send: /subscribe <topic>")
		return errors.New("invalid topic provided")
	}

	subscriptionmanager.SubscribeChatToTopic(topic, chatID)

	if err := c.Reply("topic test successful, you are now subscribed to topic"); err != nil {
		slog.Error(fmt.Sprintf("[Subscribe] [Send] %s", err.Error()))
		return err
	}

	slog.Info("[Subscribe] [Success]")
	return nil
}
