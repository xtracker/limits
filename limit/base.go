package limit

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xtracker/limits"
)

var (
	_ limits.Limit = (*baseLimit)(nil)
)

type baseLimit struct {
	sync.Mutex
	id        string
	limit     int32
	listeners []func(int)
}

func (b *baseLimit) GetLimit() int {
	return int(atomic.LoadInt32(&b.limit))
}

func (b *baseLimit) setLimit(new int) {
	if b.GetLimit() == new {
		return
	}

	atomic.StoreInt32(&b.limit, int32(new))
	b.Lock()
	defer b.Unlock()

	for _, listener := range b.listeners {
		listener(new)
	}
}

func (*baseLimit) OnSample(context.Context, time.Time, time.Duration, int, bool) {
	panic("baseLimit.OnSample must be overrided")
}

func (b *baseLimit) NotifyChange(listener func(int)) {
	b.Lock()
	defer b.Unlock()

	b.listeners = append(b.listeners, listener)
}

func (b *baseLimit) String() string {
	return b.id
}
