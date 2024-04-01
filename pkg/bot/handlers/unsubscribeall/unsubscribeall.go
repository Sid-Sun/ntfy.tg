package unsubscribeall

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	subscriptionmanager "github.com/sid-sun/ntfy.tg/pkg/subscription_manager"
	"go.uber.org/zap"
)

// Handler handles all repeat requests
func Handler(bot *tgbotapi.BotAPI, update tgbotapi.Update, logger *zap.Logger) {
	logger.Info("[UnSubscribeAll] [Attempt]")

	chatID := update.Message.Chat.ID
	consent := update.Message.CommandArguments()
	if consent != "im sure" {
		msg := tgbotapi.NewMessage(chatID, "To unsubscribe from all topics, send: /unsubscribeall im sure")
		bot.Send(msg)
		return
	}

	topics := subscriptionmanager.UnSubscribeChatFromAllTopics(chatID)
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("You are now unsubscribed from: \n- %s", strings.Join(topics, "\n- ")))
	msg.ReplyToMessageID = update.Message.MessageID

	_, err := bot.Send(msg)
	if err != nil {
		logger.Sugar().Errorf("[UnSubscribeAll] [Send] %s", err.Error())
		return
	}

	logger.Info("[UnSubscribeAll] [Success]")
}
