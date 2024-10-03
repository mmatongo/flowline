package rate

import (
	"time"

	"github.com/mmatongo/flowline/pkg/logger"
)

const (
	rateLimit       = 10
	rateLimitPeriod = 60 * time.Second
)

var requestTimestamps = make([]time.Time, 0, rateLimit)

func LimitRequest(a *logger.App) {
	currentTime := time.Now()
	if len(requestTimestamps) == rateLimit {
		timePassed := currentTime.Sub(requestTimestamps[0])
		if timePassed < rateLimitPeriod {
			sleepTime := rateLimitPeriod - timePassed
			a.Logger.Printf("rate limit reached... sleeping for %.2f seconds.", sleepTime.Seconds())
			time.Sleep(sleepTime)
		}
		requestTimestamps = requestTimestamps[1:]
	}
	requestTimestamps = append(requestTimestamps, currentTime)
}
