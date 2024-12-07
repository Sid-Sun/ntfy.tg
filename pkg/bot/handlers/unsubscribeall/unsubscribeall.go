package unsubscribeall

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	subscriptionmanager "github.com/sid-sun/ntfy.tg/pkg/subscription_manager"
	tele "gopkg.in/telebot.v4"
)

// Handler handles all repeat requests
func Handler(c tele.Context) error {
	slog.Info("[UnSubscribeAll] [Attempt]")

	chatID := c.Chat().ID
	consent := c.Message().Payload
	if consent != "im sure" {
		c.Reply("To unsubscribe from all topics, send: /unsubscribeall im sure")
		return errors.New("invalid consent")
	}

	topics := subscriptionmanager.UnSubscribeChatFromAllTopics(chatID)

	if err := c.Reply(fmt.Sprintf("You are now unsubscribed from: \n- %s", strings.Join(topics, "\n- "))); err != nil {
		slog.Error(fmt.Sprintf("[UnSubscribeAll] [Send] %s", err.Error()))
		return err
	}

	slog.Info("[UnSubscribeAll] [Success]")
	return nil
}
