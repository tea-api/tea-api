package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"tea-api/common"
	"tea-api/setting"
)

type requestTracker struct {
	mu    sync.Mutex
	times []time.Time
}

var reqMap sync.Map

func AbnormalDetection() gin.HandlerFunc {
	return func(c *gin.Context) {
		config := setting.GetAbnormalDetection()
		if !config.Enabled {
			c.Next()
			return
		}

		identifier := c.ClientIP()
		val, _ := reqMap.LoadOrStore(identifier, &requestTracker{})
		tracker := val.(*requestTracker)
		now := time.Now()

		tracker.mu.Lock()
		tracker.times = append(tracker.times, now)
		// Clean up old entries for high frequency rule (1s window)
		window := now.Add(-1 * time.Second)
		i := 0
		for ; i < len(tracker.times); i++ {
			if tracker.times[i].After(window) {
				break
			}
		}
		tracker.times = tracker.times[i:]
		count := len(tracker.times)
		tracker.mu.Unlock()

		if config.Rules.HighFrequency.Enabled &&
			count > config.Rules.HighFrequency.MaxRequestsPerSecond {
			common.SysLog(fmt.Sprintf("abnormal high frequency from %s", identifier))
			if config.Security.SleepSeconds > 0 {
				time.Sleep(time.Duration(config.Security.SleepSeconds) * time.Second)
			}
		}
		c.Next()
	}
}
