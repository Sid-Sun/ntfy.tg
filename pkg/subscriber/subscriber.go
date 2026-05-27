package subscriber

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gorilla/websocket"
	"github.com/leonklingele/passphrase"
	"github.com/sid-sun/ntfy.tg/cmd/config"
	"github.com/sid-sun/ntfy.tg/pkg/metrics"
	subscriptionmanager "github.com/sid-sun/ntfy.tg/pkg/subscription_manager"
	"go.uber.org/zap"
	tele "gopkg.in/telebot.v4"
)

type Subscriber struct {
	bot             *tele.Bot
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
		url := s.getSubscribeURL(since)
		s.logger.Sugar().Infof("[subscriber] [Subscribe] [startConnection] connection url: %s", url)
		conn, _, err = websocket.DefaultDialer.Dial(url, nil)
		if err == nil {
			conn.SetPingHandler(nil)
			s.logger.Info("[subscriber] [Subscribe] [startConnection] connected to ntfy")
			metrics.RecordWebSocketConnect()
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
				metrics.RecordWebSocketDisconnect("closed")
				return
			}

			// we want to restart here as an error here means the connection has broken
			// the message to look for is "connection timed out" but this works too, no quirks so far
			metrics.RecordWebSocketDisconnect("error")
			s.restartChan <- false
			return
		}

		metrics.RecordMessageReceived()

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
	return fmt.Sprintf("wss://%s/%s/ws?since=%d", config.GetConfig().GetNtfyDomain(), allTopics, since)
}

func (s Subscriber) sendToChats(m message) {
	start := time.Now()
	defer func() {
		metrics.ObserveForwardDuration(m.Topic, time.Since(start))
	}()

	metrics.RecordMessageForwarded(m.Topic)

	subs := subscriptionmanager.GetSubscriptions()
	var msg string
	if m.Title == "" {
		msg = fmt.Sprintf("Topic: %s \n\nMessage: %s \n", m.Topic, m.Message)
	} else {
		msg = fmt.Sprintf("Topic: %s \n\nTitle: %s \n\nMessage: %s \n", m.Topic, m.Title, m.Message)
	}
	for _, chatID := range subs[m.Topic] {
		if _, err := s.bot.Send(tele.ChatID(chatID), msg); err != nil {
			if strings.Contains(err.Error(), "bot was blocked by the user") {
				metrics.RecordForwardError(m.Topic, "blocked")
				subscriptionmanager.UnSubscribeChatFromAllTopics(chatID)
				return
			}
			metrics.RecordForwardError(m.Topic, "send_error")
			s.logger.Sugar().Errorf("[subscriber] [sendToChats] [Send] Error sending message to chat: %d, %s\n", chatID, err.Error())
		} else {
			metrics.RecordChatDelivery(m.Topic)
		}
	}
}

func NewSubscriber(bot *tele.Bot, rc chan bool, logger *zap.Logger) Subscriber {
	return Subscriber{
		bot:         bot,
		restartChan: rc,
		logger:      logger,
	}
}

func (s Subscriber) informAdmin(text string) {
	s.bot.Send(tele.ChatID(config.GetConfig().GetAdminChatID()), text)
}
