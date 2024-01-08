package subscriber

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	subscriptionmanager "github.com/sid-sun/ntfy.tg/pkg/subscription_manager"
)

type Subscriber struct {
	bot         *tgbotapi.BotAPI
	restartChan chan bool
}

func (s Subscriber) Subscribe() {
	fmt.Println("Subscribing to ntfy")
	defer fmt.Println("Unsubscribing from ntfy")
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
		// fmt.Println(scanner.Text())
		var m message
		err := json.Unmarshal(scanner.Bytes(), &m)
		if err != nil {
			panic(err)
			// handle error
		}
		// fmt.Printf("%v\n", m)
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
		msg := tgbotapi.NewMessage(chatID, m.Message)
		s.bot.Send(msg)
	}
}

func NewSubscriber(bot *tgbotapi.BotAPI, rc chan bool) Subscriber {
	return Subscriber{
		bot:         bot,
		restartChan: rc,
	}
}
