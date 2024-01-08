package subscriptionmanager

var subscriptions map[string][]int64
var restartChan chan bool

func SubscribeChatToTopic(topic string, chatID int64) {
	if subscriptions[topic] == nil {
		subscriptions[topic] = []int64{chatID}
		restartChan <- true
		return
	}
	subscriptions[topic] = append(subscriptions[topic], chatID)
}

func GetSubscriptions() map[string][]int64 {
	return subscriptions
}

func InitSubscriptions(rsc chan bool) {
	restartChan = rsc
	// fetch from store
	subscriptions = make(map[string][]int64)
	// Initial subscriptions to prevent panic when no subscriptions are found
	// TODO: remove this when we have a real store
	subscriptions["verysecrettopic_7378273298273298"] = []int64{191332017}
}
