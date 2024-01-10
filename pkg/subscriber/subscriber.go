package subscriber

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/leonklingele/passphrase"
	"github.com/sid-sun/ntfy.tg/cmd/config"
	subscriptionmanager "github.com/sid-sun/ntfy.tg/pkg/subscription_manager"
	"go.uber.org/zap"
)

type Subscriber struct {
	bot         *tgbotapi.BotAPI
	restartChan chan bool
	logger      *zap.Logger
}

func getRandomName() string {
	passphrase.Separator = "-"
	phrase, _ := passphrase.Generate(2)
	return phrase
}

func (s Subscriber) Subscribe() {

	name := getRandomName()
	s.informAdmin(fmt.Sprintf("Subscribing to ntfy [%s]", name))
	s.logger.Sugar().Infof("Subscribing to ntfy [%s]\n", name)
	defer func() {
		s.informAdmin(fmt.Sprintf("Unsubscribing from ntfy [%s]", name))
		s.logger.Sugar().Infof("Unsubscribing from ntfy [%s]\n", name)
	}()

	resp, err := http.Get(s.getSubscribeURL())
	if err != nil {
		return
	}
	defer resp.Body.Close()
	go func() {
		<-s.restartChan
		resp.Body.Close()
		s.Subscribe()
	}()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		var m message
		err := json.Unmarshal(scanner.Bytes(), &m)
		if err != nil {
			panic(err)
		}
		if m.Event == "message" {
			go s.sendToChats(m)
		}
	}
}

func (s Subscriber) getSubscribeURL() string {
	topics := []string{}
	for topic := range subscriptionmanager.GetSubscriptions() {
		topics = append(topics, topic)
	}
	allTopics := strings.Join(topics, ",")
	return fmt.Sprintf("https://ntfy.sh/%s/json", allTopics)
}

func (s Subscriber) sendToChats(m message) {
	subs := subscriptionmanager.GetSubscriptions()
	for _, chatID := range subs[m.Topic] {
		var msg tgbotapi.MessageConfig
		if m.Title == "" {
			msg = tgbotapi.NewMessage(chatID, fmt.Sprintf("_%s_\n\n`%s`\n", m.Topic, m.Message))
		} else {
			msg = tgbotapi.NewMessage(chatID, fmt.Sprintf("_%s_\n\n*%s*\n\n `%s`\n", m.Topic, m.Title, m.Message))
		}
		msg.ParseMode = "Markdown"
		s.bot.Send(msg)
	}
}

func NewSubscriber(bot *tgbotapi.BotAPI, rc chan bool, logger *zap.Logger) Subscriber {
	return Subscriber{
		bot:         bot,
		restartChan: rc,
		logger:      logger,
	}
}

func (s Subscriber) informAdmin(text string) {
	msg := tgbotapi.NewMessage(config.GetConfig().GetAdminChatID(), text)
	s.bot.Send(msg)
}
