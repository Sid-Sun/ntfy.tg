package metrics

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// StartServer starts an HTTP server that exposes Prometheus metrics on /metrics
// and a health check on /healthz. It blocks, so call it in a goroutine.
func StartServer(port string, logger *zap.Logger) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	addr := fmt.Sprintf(":%s", port)
	logger.Info(fmt.Sprintf("[Metrics] Starting metrics server on %s", addr))
	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Fatal(fmt.Sprintf("[Metrics] Metrics server failed: %s", err.Error()))
	}
}
