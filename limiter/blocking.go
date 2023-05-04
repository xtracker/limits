package limiter

import (
	"context"
	"errors"
	"time"

	"github.com/xtracker/limits"
)

type BlockingLimiter struct {
	limits.Limiter
	timeout time.Duration
	ch      chan struct{}
}

func (b *BlockingLimiter) Acquire(ctx context.Context) (limits.Listener, error) {
	listener, err := b.tryAcquire(ctx)

	if err == nil {
		return listener, nil
	}

	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	for {
		select {
		case <-b.ch:

		case <-ctx.Done():
			return nil, errors.New("timeout")
		}

		listener, err := b.tryAcquire(ctx)

		if err == nil {
			return listener, nil
		}
	}

}

func (b *BlockingLimiter) tryAcquire(ctx context.Context) (limits.Listener, error) {
	listener, err := b.Limiter.Acquire(ctx)
	if err == nil {
		return func(ctx context.Context, result limits.Result) {
			listener(ctx, result)
			select {
			case b.ch <- struct{}{}:
			default:
			}
		}, nil
	}

	return nil, err
}
