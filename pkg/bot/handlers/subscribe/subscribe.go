package subscribe

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
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
	if len(args) != 1 {
		msg := tgbotapi.NewMessage(chatID, "invalid message, to subscribe send: /subscribe <topic>")
		bot.Send(msg)
		return
	}
	topic := args[0]
	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	randomMessage := base64.StdEncoding.EncodeToString(randomBytes)
	testChannel := make(chan bool, 1)

	testTopic := strings.Join([]string{topic, "test"}, "_")
	go testReceive(testTopic, randomMessage, testChannel)
	resp, err := http.Post(fmt.Sprintf("https://ntfy.sh/%s", testTopic), "text/plain", strings.NewReader(randomMessage))
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

	// body, _ := io.ReadAll(resp.Body)
	// var publishedMessage publishedMessage
	// _ = json.Unmarshal(body, &publishedMessage)
	// testReceive(publishedMessage)
	<-testChannel

	subscriptionmanager.SubscribeChatToTopic(topic, chatID)
	msg := tgbotapi.NewMessage(chatID, "topic test successful, you are now subscribed to topic")
	msg.ReplyToMessageID = update.Message.MessageID

	_, err = bot.Send(msg)
	if err != nil {
		logger.Sugar().Errorf("[%s] [%s] %s", handler, "Send", err.Error())
		return
	}

	logger.Info("[Subscribe] [Success]")
}

func testReceive(topic, message string, c chan bool) {
	resp, err := http.Get(fmt.Sprintf("https://ntfy.sh/%s/raw", topic))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		// fmt.Println(fmt.Sprintf("Random message: %s Received message: %s", message, scanner.Text()))
		if message == scanner.Text() {
			c <- true
			return
		}
	}
}
