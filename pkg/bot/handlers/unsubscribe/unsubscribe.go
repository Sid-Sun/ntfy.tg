package unsubscribe

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"

	subscriptionmanager "github.com/sid-sun/ntfy.tg/pkg/subscription_manager"
	tele "gopkg.in/telebot.v4"
)

// Handler handles all repeat requests
func Handler(c tele.Context) error {
	slog.Info("[UnSubscribe] [Attempt]")

	chatID := c.Chat().ID
	topic := c.Message().Payload
	if topic == "" {
		c.Reply("Invalid message, to unsubscribe send: /unsubscribe <topic>")
		return errors.New("no topic provided")
	}

	subTopics := subscriptionmanager.GetChatSubscriptions(chatID)
	if len(subTopics) == 0 {
		msg := "You are not subscribed to any topics, to subscribe to a topic use /subscribe <topic>"
		if err := c.Reply(msg); err != nil {
			slog.Error(fmt.Sprintf("[Unsubscribeall] [GetChatSubscriptions] [noTopics] [Send] %s", err.Error()))
			return err
		}
		return nil
	}
	if !slices.Contains(subTopics, topic) {
		msg := "You are not subscribed to this topic, to subscribe to a topic use /subscribe <topic>"
		if err := c.Reply(msg); err != nil {
			slog.Error(fmt.Sprintf("[Unsubscribeall] [GetChatSubscriptions] [notSub] [Send] %s", err.Error()))
			return err
		}
		return nil
	}

	subscriptionmanager.UnSubscribeChatToTopic(topic, chatID)

	if err := c.Send(fmt.Sprintf("You are now unsubscribed from %s", topic)); err != nil {
		slog.Error(fmt.Sprintf("[UnSubscribe] [Send] %s", err.Error()))
		return err
	}

	slog.Info("[UnSubscribe] [Success]")
	return nil
}
