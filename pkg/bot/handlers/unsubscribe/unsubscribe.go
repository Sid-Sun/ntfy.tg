package unsubscribe

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	subscriptionmanager "github.com/sid-sun/ntfy.tg/pkg/subscription_manager"
	"go.uber.org/zap"
)

// Handler handles all repeat requests
func Handler(bot *tgbotapi.BotAPI, update tgbotapi.Update, logger *zap.Logger) {
	logger.Info("[UnSubscribe] [Attempt]")

	chatID := update.Message.Chat.ID
	var topic string
	args := strings.Split(update.Message.CommandArguments(), " ")
	topic = args[0]
	if len(args) != 1 || topic == "" {
		msg := tgbotapi.NewMessage(chatID, "Invalid message, to unsubscribe send: /unsubscribe <topic>")
		bot.Send(msg)
		return
	}

	subscriptionmanager.UnSubscribeChatToTopic(topic, chatID)
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("You are now unsubscribed from %s", topic))
	msg.ReplyToMessageID = update.Message.MessageID

	_, err := bot.Send(msg)
	if err != nil {
		logger.Sugar().Errorf("[UnSubscribe] [Send] %s", err.Error())
		return
	}

	logger.Info("[UnSubscribe] [Success]")
}
