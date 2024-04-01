package monitoring

import (
	"net/http"
	"time"

	"github.com/sid-sun/ntfy.tg/cmd/config"
	"go.uber.org/zap"
)

func PeriodicNotify(logger *zap.Logger) {
	urls := config.GetConfig().PingURLs
	if len(urls) <= 1 {
		logger.Info("[Monitoring] [PeriodicNotify] no URLs to ping, quitting periodic notify")
		return
	}
	for range time.Tick(time.Minute) {
		for _, url := range urls {
			if url == "" {
				continue
			}
			res, err := http.Get(url)
			if err != nil {
				logger.Sugar().Errorf("[Monitoring] [PeriodicNotify] failed to ping %s - error: %s", url, err.Error())
				continue
			}
			if res.StatusCode != http.StatusOK {
				logger.Sugar().Errorf("[Monitoring] [PeriodicNotify] ping %s status code not OK: %d", url, res.StatusCode)
			}
		}
	}
}
