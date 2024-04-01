package subscriber

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gorilla/websocket"
	"github.com/leonklingele/passphrase"
	"github.com/sid-sun/ntfy.tg/cmd/config"
	subscriptionmanager "github.com/sid-sun/ntfy.tg/pkg/subscription_manager"
	"go.uber.org/zap"
)

type Subscriber struct {
	bot             *tgbotapi.BotAPI
	restartChan     chan bool
	logger          *zap.Logger
	lastMessageTime int64
}

func getRandomName() string {
	passphrase.Separator = "-"
	phrase, _ := passphrase.Generate(2)
	return phrase
}

func (s Subscriber) Subscribe() {
	var conn *websocket.Conn
	var err error
	startConnection := func() error {
		since := s.lastMessageTime
		if since == 0 {
			since = time.Now().Unix()
		}
		conn, _, err = websocket.DefaultDialer.Dial(s.getSubscribeURL(since), nil)
		if err == nil {
			conn.SetPingHandler(nil)
			s.logger.Info("[subscriber] [Subscribe] [startConnection] connected to ntfy")
			go s.listenForMessages(conn)
			return nil
		}
		if err != nil {
			s.logger.Sugar().Errorf("[subscriber] [Subscribe] [startConnection] error connecting to ntfy: %s", err.Error())
			return err
		}
		return nil
	}

	err = backoff.Retry(startConnection, backoff.NewExponentialBackOff())
	if err != nil {
		s.logger.Sugar().Errorf("exponential backoff retry error: %s", err.Error())
		return
	}

	defer conn.Close()

	for {
		<-s.restartChan
		s.logger.Info("[subscriber] [Subscribe] Restarting connection - restart signal received")
		conn.Close()

		// defer conn.Close() is not required here
		// as defer is already called and only underlying resp will change
		err := backoff.Retry(startConnection, backoff.NewExponentialBackOff())
		if err != nil {
			s.logger.Sugar().Errorf("exponential backoff retry error: %s", err.Error())
			return
		}
	}
}

func (s Subscriber) listenForMessages(conn *websocket.Conn) {
	name := getRandomName()
	s.informAdmin(fmt.Sprintf("Subscribing to ntfy [%s]", name))
	s.logger.Sugar().Infof("[subscriber] [listenForMessages] Subscribing to ntfy [%s]\n", name)
	defer func() {
		// spin informAdmin in a different routine as this is a blocking call to tg api library
		// and if we are exiting due to connection reset / timeout
		// that call would otherwise block return in cases of issues on our side
		go s.informAdmin(fmt.Sprintf("Unsubscribing from ntfy [%s]", name))
		s.logger.Sugar().Infof("[subscriber] [listenForMessages] Unsubscribing from ntfy [%s]\n", name)
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			s.logger.Sugar().Errorf("Error reading message:", err)

			// this error is thrown when conn is closed by Subscribe for a restart
			// if it is not handled, a restart loop is triggered
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}

			// we want to restart here as an error here means the connection has broken
			// the message to look for is "connection timed out" but this works too, no quirks so far
			s.restartChan <- false
			return
		}

		// s.logger.Sugar().Infof("[subscriber] [listener] Received a new message, time since last message: %d\n", (time.Now().Unix() - s.lastMessageTime))

		// handle message
		var m message
		err = json.Unmarshal(msg, &m)
		if err != nil {
			panic(err)
		}
		if m.Event == "message" {
			go s.sendToChats(m)
		}
		s.lastMessageTime = m.Time
	}
}

func (s Subscriber) getSubscribeURL(since int64) string {
	topics := []string{}
	for topic := range subscriptionmanager.GetSubscriptions() {
		topics = append(topics, topic)
	}
	allTopics := strings.Join(topics, ",")
	return fmt.Sprintf("wss://ntfy.sh/%s/ws?since=%d", allTopics, since)
}

func (s Subscriber) sendToChats(m message) {
	subs := subscriptionmanager.GetSubscriptions()
	for _, chatID := range subs[m.Topic] {
		var msg tgbotapi.MessageConfig
		if m.Title == "" {
			msg = tgbotapi.NewMessage(chatID, fmt.Sprintf("Topic: %s \n\nMessage: %s \n", m.Topic, m.Message))
		} else {
			msg = tgbotapi.NewMessage(chatID, fmt.Sprintf("Topic: %s \n\nTitle: %s \n\nMessage: %s \n", m.Topic, m.Title, m.Message))
		}
		_, err := s.bot.Send(msg)
		if err != nil {
			s.logger.Sugar().Errorf("[subscriber] [sendToChats] [Send] Error sending message to chat: %d, %s\n", chatID, err.Error())
		}
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
