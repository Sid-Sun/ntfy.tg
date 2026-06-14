package subscribe

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/leonklingele/passphrase"
	"github.com/sid-sun/ntfy.tg/cmd/config"
	"github.com/sid-sun/ntfy.tg/pkg/metrics"
	subscriptionmanager "github.com/sid-sun/ntfy.tg/pkg/subscription_manager"
	tele "gopkg.in/telebot.v4"
)

// Handler handles all repeat requests
func Hnadler(c tele.Context) error {
	metrics.RecordBotCommand("subscribe")
	slog.Info("[Subscribe] [Attempt]")

	chatID := c.Chat().ID
	topic := c.Message().Payload
	if topic == "" {
		c.Reply("invalid message, to subscribe send: /subscribe <topic>")
		return errors.New("no topic provided")
	}

	validTopic := allowedTopicRegex.MatchString(topic)
	if !validTopic {
		c.Reply("invalid topic, to subscribe send: /subscribe <topic>")
		return errors.New("invalid topic provided")
	}

	if err := testTopic(topic); err != nil {
		slog.Error(fmt.Sprintf("[Subscribe] [testTopic] %s", err.Error()))
		c.Reply(fmt.Sprintf("topic test failed: %s", err.Error()))
		return err
	}

	subscriptionmanager.SubscribeChatToTopic(topic, chatID)

	if err := c.Reply("topic test successful, you are now subscribed to topic"); err != nil {
		slog.Error(fmt.Sprintf("[Subscribe] [Send] %s", err.Error()))
		return err
	}

	slog.Info("[Subscribe] [Success]")
	return nil
}

func testTopic(topic string) error {
	passphrase.Separator = "-"
	randomMessage, err := passphrase.Generate(4)
	if err != nil {
		return fmt.Errorf("generating test message: %w", err)
	}

	domain := config.GetConfig().GetNtfyDomain()
	wsURL := fmt.Sprintf("wss://%s/%s/ws?since=%d", domain, topic, time.Now().Unix())

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("connecting to topic: %w", err)
	}
	defer conn.Close()

	result := make(chan error, 1)
	go func() {
		conn.SetReadDeadline(time.Now().Add(15 * time.Second))
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				result <- fmt.Errorf("waiting for test message: %w", err)
				return
			}
			var m publishedMessage
			if err := json.Unmarshal(msg, &m); err != nil {
				continue
			}
			if m.Event == "message" && m.Message == randomMessage {
				result <- nil
				return
			}
		}
	}()

	postURL := fmt.Sprintf("https://%s/%s", domain, topic)
	req, _ := http.NewRequest(http.MethodPost, postURL, strings.NewReader(randomMessage))
	req.Header.Set("Content-Type", "text/plain")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("publishing test message: %w", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("publishing test message: HTTP %d", resp.StatusCode)
	}

	return <-result
}
