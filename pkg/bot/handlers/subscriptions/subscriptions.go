package subscriptions

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	subscriptionmanager "github.com/sid-sun/ntfy.tg/pkg/subscription_manager"
	"go.uber.org/zap"
)

// Handler handles all repeat requests
func Handler(bot *tgbotapi.BotAPI, update tgbotapi.Update, logger *zap.Logger) {
	logger.Info("[Subscriptions] [Attempt]")

	chatID := update.Message.Chat.ID

	// Get chat subscriptions and send them to the user
	topics := subscriptionmanager.GetChatSubscriptions(chatID)
	var msg tgbotapi.MessageConfig
	if len(topics) == 0 {
		msg = tgbotapi.NewMessage(chatID, "You are not subscribed to any topics, to subscribe to a topic use /subscribe <topic>")
	} else {
		msg = tgbotapi.NewMessage(chatID, fmt.Sprintf("You are subscribed to: \n- %s", strings.Join(topics, "\n- ")))
	}
	msg.ReplyToMessageID = update.Message.MessageID

	_, err := bot.Send(msg)
	if err != nil {
		logger.Sugar().Errorf("[Subscriptions] [Send] %s", err.Error())
		return
	}

	logger.Info("[Subscriptions] [Success]")
}
