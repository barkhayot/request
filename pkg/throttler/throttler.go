package throttler

import (
	"context"

	"golang.org/x/time/rate"
)

type Throttler interface {
	Wait(ctx context.Context) error
}

type RateThrottler struct {
	limiter *rate.Limiter
}

func NewRateThrottler(r rate.Limit, burst int) *RateThrottler {
	return &RateThrottler{
		limiter: rate.NewLimiter(r, burst),
	}
}

func (t *RateThrottler) Wait(ctx context.Context) error {
	return t.limiter.Wait(ctx)
}
