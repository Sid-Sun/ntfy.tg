package subscriber

// from https://github.com/binwiederhier/ntfy
type message struct { // TODO combine with server.message
	ID      string
	Event   string
	Time    int64
	Topic   string
	Message string
	Title   string
}
