package limits

import (
	"context"
	"time"
)

type Limit interface {
	GetLimit() int
	OnSample(ctx context.Context, startTime time.Time, rtt time.Duration, inflight int, dropped bool)
	NotityChange(func(int))
	String() string
}
