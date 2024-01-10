package subscribe

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/leonklingele/passphrase"
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

	testTopic(topic, chatID, bot)

	subscriptionmanager.SubscribeChatToTopic(topic, chatID)
	msg := tgbotapi.NewMessage(chatID, "topic test successful, you are now subscribed to topic")
	msg.ReplyToMessageID = update.Message.MessageID

	_, err := bot.Send(msg)
	if err != nil {
		logger.Sugar().Errorf("[%s] [%s] %s", handler, "Send", err.Error())
		return
	}

	logger.Info("[Subscribe] [Success]")
}

func testTopic(topic string, chatID int64, bot *tgbotapi.BotAPI) {
	randomMessage := getRandomValue()
	testChannel := make(chan bool, 1)

	// create test topic and subscribe to it
	// once initial message is received, send signal to start test
	// once test message is received, send signal to stop test
	testTopicName := strings.Join([]string{topic, "test"}, "_")
	go testReceive(testTopicName, randomMessage, testChannel)
	<-testChannel

	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("https://ntfy.sh/%s", testTopicName), strings.NewReader(randomMessage))
	req.Header.Set("Content-Type", "text/plain")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("something went wrong: %s", err.Error()))
		bot.Send(msg)
		return
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("expected HTTP Status 200, got: %s %d", string(body), resp.StatusCode))
		bot.Send(msg)
		return
	}

	// wait for test message to be checked successfully
	<-testChannel
}

func testReceive(topic, message string, c chan bool) {
	resp, err := http.Get(fmt.Sprintf("https://ntfy.sh/%s/raw", topic))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		// send singal to start test -> initial message.
		c <- true
		if message == scanner.Text() {
			c <- true
			return
		}
	}
}

func getRandomValue() string {
	passphrase.Separator = "-"
	phrase, _ := passphrase.Generate(4)
	return phrase
}
