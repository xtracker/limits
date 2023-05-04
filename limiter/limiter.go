package limiter

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/xtracker/limits"
)

var (
	_ limits.Limiter = (*simpleLimiter)(nil)
)

var errLimitExceeded = errors.New("limits error: max inflight exceeded")

func NewSimpleLimiter(id string, limitAlgorithm limits.Limit) limits.Limiter {
	return &simpleLimiter{
		id:             id,
		limitAlgorithm: limitAlgorithm,
	}
}

type simpleLimiter struct {
	id             string
	limitAlgorithm limits.Limit
	inFlight       int32
}

func (l *simpleLimiter) getInFlight() int {
	return int(atomic.LoadInt32(&l.inFlight))
}

func (l *simpleLimiter) createListener() limits.Listener {
	startTime := time.Now()
	inFlight := int(atomic.AddInt32(&l.inFlight, 1))
	return func(ctx context.Context, result limits.Result) {
		switch result {
		case limits.SUCCESS:
			l.limitAlgorithm.OnSample(ctx, startTime, time.Since(startTime), inFlight, false)
		case limits.DROPPED:
			l.limitAlgorithm.OnSample(ctx, startTime, time.Since(startTime), inFlight, true)
		case limits.IGNORED:
		default:
			// nerver reached path
		}
	}
}

func (l *simpleLimiter) Acquire(ctx context.Context) (limits.Listener, error) {
	inFlight := l.getInFlight()
	if inFlight < l.limitAlgorithm.GetLimit() {
		return l.createListener(), nil
	}

	return nil, errLimitExceeded
}
