package subscribe

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	subscriptionmanager "github.com/sid-sun/ntfy.tg/pkg/subscription_manager"
	"go.uber.org/zap"
)

// Handler handles all repeat requests
func Handler(bot *tgbotapi.BotAPI, update tgbotapi.Update, logger *zap.Logger) {
	logger.Info("[Subscribe] [Attempt]")

	chatID := update.Message.Chat.ID
	args := strings.Split(update.Message.CommandArguments(), " ")
	topic := args[0]
	if len(args) != 1 || topic == "" {
		msg := tgbotapi.NewMessage(chatID, "invalid message, to subscribe send: /subscribe <topic>")
		bot.Send(msg)
		return
	}

	validTopic := allowedTopicRegex.MatchString(topic)
	if !validTopic {
		msg := tgbotapi.NewMessage(chatID, "invalid topic, to subscribe send: /subscribe <topic>")
		bot.Send(msg)
		return
	}

	subscriptionmanager.SubscribeChatToTopic(topic, chatID)
	msg := tgbotapi.NewMessage(chatID, "topic test successful, you are now subscribed to topic")
	msg.ReplyToMessageID = update.Message.MessageID

	_, err := bot.Send(msg)
	if err != nil {
		logger.Sugar().Errorf("[Subscribe] [Send] %s", err.Error())
		return
	}

	logger.Info("[Subscribe] [Success]")
}
