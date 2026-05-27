package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var startTime = time.Now()

// --- Message Forwarding (per-topic) ---

var messagesForwardedByTopic = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "ntfy_tg_messages_forwarded_total",
	Help: "Total messages forwarded from ntfy to Telegram, by topic.",
}, []string{"topic"})

var chatDeliveriesByTopic = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "ntfy_tg_chat_deliveries_total",
	Help: "Total individual chat deliveries (one message to N chats = N increments), by topic.",
}, []string{"topic"})

var forwardErrorsByTopic = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "ntfy_tg_message_forward_errors_total",
	Help: "Total failed message forwards, by topic and reason.",
}, []string{"topic", "reason"})

var forwardDurationByTopic = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "ntfy_tg_message_forward_duration_seconds",
	Help:    "Time to forward a message to all subscribed chats, by topic.",
	Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
}, []string{"topic"})

// --- Message Forwarding (aggregate) ---

var messagesForwardedAggregate = promauto.NewCounter(prometheus.CounterOpts{
	Name: "ntfy_tg_messages_forwarded_aggregate_total",
	Help: "Total messages forwarded from ntfy to Telegram (aggregate across all topics).",
})

var chatDeliveriesAggregate = promauto.NewCounter(prometheus.CounterOpts{
	Name: "ntfy_tg_chat_deliveries_aggregate_total",
	Help: "Total individual chat deliveries across all topics.",
})

var forwardErrorsAggregate = promauto.NewCounter(prometheus.CounterOpts{
	Name: "ntfy_tg_message_forward_errors_aggregate_total",
	Help: "Total failed message forwards across all topics.",
})

var forwardDurationAggregate = promauto.NewHistogram(prometheus.HistogramOpts{
	Name:    "ntfy_tg_message_forward_duration_aggregate_seconds",
	Help:    "Time to forward a message to all subscribed chats (aggregate).",
	Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
})

// --- Subscription Gauges ---

var topicsActive = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "ntfy_tg_topics_active",
	Help: "Current number of active topics with at least one subscriber.",
})

var usersActive = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "ntfy_tg_users_active",
	Help: "Current number of unique subscribed Telegram chat IDs.",
})

var subscriptionsActive = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "ntfy_tg_subscriptions_active",
	Help: "Total active topic-user subscription pairs.",
})

// --- Bot Commands ---

var botCommands = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "ntfy_tg_bot_commands_total",
	Help: "Total bot commands received, by command name.",
}, []string{"command"})

// --- WebSocket ---

var wsDisconnectionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "ntfy_tg_websocket_disconnections_total",
	Help: "Total WebSocket disconnections, by reason.",
}, []string{"reason"})

var wsConnected = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "ntfy_tg_websocket_connected",
	Help: "Whether the bot is currently connected to ntfy via WebSocket (1 = connected, 0 = disconnected).",
})

var messagesReceived = promauto.NewCounter(prometheus.CounterOpts{
	Name: "ntfy_tg_messages_received_total",
	Help: "Total raw messages received from the ntfy WebSocket (all event types).",
})

// --- Uptime ---

var uptimeSeconds = promauto.NewGaugeFunc(prometheus.GaugeOpts{
	Name: "ntfy_tg_uptime_seconds",
	Help: "Process uptime in seconds.",
}, func() float64 {
	return time.Since(startTime).Seconds()
})

// Ensure uptimeSeconds is referenced to avoid unused variable lint error.
var _ = uptimeSeconds

// --- Public recording functions ---

// RecordMessageForwarded increments both per-topic and aggregate message forwarded counters.
func RecordMessageForwarded(topic string) {
	messagesForwardedByTopic.WithLabelValues(topic).Inc()
	messagesForwardedAggregate.Inc()
}

// RecordChatDelivery increments both per-topic and aggregate chat delivery counters.
func RecordChatDelivery(topic string) {
	chatDeliveriesByTopic.WithLabelValues(topic).Inc()
	chatDeliveriesAggregate.Inc()
}

// RecordForwardError increments both per-topic and aggregate forward error counters.
func RecordForwardError(topic string, reason string) {
	forwardErrorsByTopic.WithLabelValues(topic, reason).Inc()
	forwardErrorsAggregate.Inc()
}

// ObserveForwardDuration records the time taken to forward a message, in both per-topic and aggregate histograms.
func ObserveForwardDuration(topic string, duration time.Duration) {
	seconds := duration.Seconds()
	forwardDurationByTopic.WithLabelValues(topic).Observe(seconds)
	forwardDurationAggregate.Observe(seconds)
}

// RecordBotCommand increments the bot command counter for the given command.
func RecordBotCommand(command string) {
	botCommands.WithLabelValues(command).Inc()
}

// RecordWebSocketConnect sets the connected gauge to 1.
func RecordWebSocketConnect() {
	wsConnected.Set(1)
}

// RecordWebSocketDisconnect increments the disconnection counter and sets the connected gauge to 0.
func RecordWebSocketDisconnect(reason string) {
	wsDisconnectionsTotal.WithLabelValues(reason).Inc()
	wsConnected.Set(0)
}

// SetWebSocketConnected sets the connected gauge directly.
func SetWebSocketConnected(connected bool) {
	if connected {
		wsConnected.Set(1)
	} else {
		wsConnected.Set(0)
	}
}

// RecordMessageReceived increments the raw messages received counter.
func RecordMessageReceived() {
	messagesReceived.Inc()
}

// UpdateSubscriptionGauges sets the current topic, user, and subscription counts.
func UpdateSubscriptionGauges(topics, users, subscriptions int) {
	topicsActive.Set(float64(topics))
	usersActive.Set(float64(users))
	subscriptionsActive.Set(float64(subscriptions))
}

func init() {
	// Initialize known bot commands to 0
	for _, cmd := range []string{"subscribe", "unsubscribe", "unsubscribeall", "subscriptions", "start", "help"} {
		botCommands.WithLabelValues(cmd)
	}
	// Initialize known websocket disconnection reasons to 0
	wsDisconnectionsTotal.WithLabelValues("error")
	wsDisconnectionsTotal.WithLabelValues("closed")
}

// InitTopicMetrics initializes the metrics for a given topic to 0.
// This is critical so that Prometheus's increase() function can correctly see
// the transition from 0 to 1 for low-traffic topics.
func InitTopicMetrics(topic string) {
	messagesForwardedByTopic.WithLabelValues(topic)
	chatDeliveriesByTopic.WithLabelValues(topic)
	forwardErrorsByTopic.WithLabelValues(topic, "blocked")
	forwardErrorsByTopic.WithLabelValues(topic, "send_error")
	forwardDurationByTopic.WithLabelValues(topic)
}
