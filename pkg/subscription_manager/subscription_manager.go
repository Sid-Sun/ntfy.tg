package subscriptionmanager

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"sync"

	"github.com/fitant/storage-engine-go/storageengine"
	"github.com/sid-sun/ntfy.tg/cmd/config"
)

var subscriptionsObject *storageengine.Object
var subscriptionsMutex *sync.Mutex
var subscriptions map[string][]int64
var restartChan chan bool

func SubscribeChatToTopic(topic string, chatID int64) {
	if subscriptions[topic] == nil {
		subscriptionsMutex.Lock()
		subscriptions[topic] = []int64{chatID}
		subscriptionsMutex.Unlock()
		saveToSE()
		return
	}
	for _, id := range subscriptions[topic] {
		if id == chatID {
			return
		}
	}
	subscriptions[topic] = append(subscriptions[topic], chatID)
}

func GetChatSubscriptions(chatID int64) []string {
	var topics []string
	for topic, chats := range subscriptions {
		for _, id := range chats {
			if id == chatID {
				topics = append(topics, topic)
			}
		}
	}
	return topics
}

func UnSubscribeChatFromAllTopics(chatID int64) []string {
	var changed bool
	subscriptionsMutex.Lock()
	var unsubscribedTopics []string
	for topic, chats := range subscriptions {
		subscriptions[topic] = slices.DeleteFunc(chats, func(n int64) bool {
			if n == chatID {
				unsubscribedTopics = append(unsubscribedTopics, topic)
				changed = true
				return true
			}
			return false
		})
		if len(subscriptions[topic]) == 1 && subscriptions[topic][0] == 0 {
			delete(subscriptions, topic)
		}
	}
	subscriptionsMutex.Unlock()
	if changed {
		saveToSE()
	}
	return unsubscribedTopics
}

func UnSubscribeChatToTopic(topic string, chatID int64) {
	if subscriptions[topic] == nil {
		return
	}
	var changed bool
	subscriptionsMutex.Lock()
	subscriptions[topic] = slices.DeleteFunc(subscriptions[topic], func(n int64) bool {
		if n == chatID {
			changed = true
			return true
		}
		return false
	})
	if changed && len(subscriptions[topic]) == 1 && subscriptions[topic][0] == 0 {
		delete(subscriptions, topic)
	}
	subscriptionsMutex.Unlock()
	if changed {
		saveToSE()
	}
}

func GetSubscriptions() map[string][]int64 {
	return subscriptions
}

func InitSubscriptions(rsc chan bool) {
	subscriptions = make(map[string][]int64)
	loadDataFromSE()
	restartChan = rsc
	// fetch from store
	// Initial subscriptions to prevent panic when no subscriptions are found
	subscriptions["verysecrettopic_7378273298273298"] = []int64{config.GetConfig().GetAdminChatID()}
	// for topic, chats := range subscriptions {
	// 	for _, chatID := range chats {
	// 		log.Printf("Subscribed to %s: %d\n", topic, chatID)
	// 	}
	// }
}

func saveToSE() {
	subscriptionsMutex.Lock()
	defer subscriptionsMutex.Unlock()
	data, err := json.Marshal(subscriptions)
	if err != nil {
		panic(err)
	}
	err = subscriptionsObject.SetData(string(data))
	if err != nil {
		panic(err)
	}
	err = subscriptionsObject.Publish()
	if err != nil {
		panic(err)
	}
	restartChan <- true
}

func loadDataFromSE() {
	subscriptionsMutex = new(sync.Mutex)
	subscriptionsMutex.Lock()
	defer subscriptionsMutex.Unlock()
	// fetch from store
	cfg := config.GetConfig().StorageEngine
	if seClient, err := storageengine.NewClientConfig(http.DefaultClient, cfg.URL); err != nil {
		panic(err)
	} else {
		subscriptionsObject, err = storageengine.NewObject(seClient)
		if err != nil {
			panic(err)
		}
	}

	subscriptionsObject.SetID(cfg.ObjectID)
	subscriptionsObject.SetPassword(cfg.ObjectPassword)
	if err := subscriptionsObject.Refresh(); err != nil {
		log.Print(err)
	}
	if subscriptionsObject.GetData() == "" {
		log.Print("fetch from SE failed presumably due to 404 - doing a fresh start")
	} else {
		err := json.Unmarshal([]byte(subscriptionsObject.GetData()), &subscriptions)
		if err != nil {
			log.Print(err)
		}
	}
}
