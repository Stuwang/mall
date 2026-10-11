package orderbiz

import (
	"time"

	"helloworld/base/logger"
)

// StartTimeoutWorker 启动支付超时扫描
func StartTimeoutWorker(interval time.Duration) {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			CancelExpiredOrders()
		}
	}()
	logger.Infof("order timeout worker started interval=%s", interval)
}
