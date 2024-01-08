package subscribe

type publishedMessage struct {
	Id      string
	Time    int64
	Expires int64
	Event   string
	Topic   string
	Message string
}
