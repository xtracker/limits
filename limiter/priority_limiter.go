package limiter

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xtracker/limits"
	"github.com/xtracker/limits/util"
)

var (
	_                  limits.Limiter = (*priorityLimiter)(nil)
	errBacklogOverload                = errors.New("backlog overload")
	errEvicted                        = errors.New("evicted by higher priority")
	errTimeout                        = errors.New("wait timeout")
)

type eventData struct {
	listener limits.Listener
	err      error
}

type event struct {
	priority int
	c        chan eventData
	ctx      context.Context
	fifo     bool
	done     int32
}

func (e *event) Done() bool {
	return atomic.LoadInt32(&e.done) == 1
}

func (e *event) cancel() {
	atomic.StoreInt32(&e.done, 1)
}

func (e *event) Less(other util.Comparable) bool {
	oe := other.(*event)
	di, _ := e.ctx.Deadline()
	dj, _ := oe.ctx.Deadline()

	// fifo mode:
	if e.fifo {
		return di.Before(dj)
	}

	// priority & lifo mode
	if e.priority == oe.priority {
		return di.After(dj)
	}

	return e.priority > oe.priority
}

func (e *event) signal(listener limits.Listener, err error) bool {
	// use unbuffered ch to make sure the listener is consumed
	// 1. when context is done, return false to indicate that signal failed
	// 2. or it is soon to be consumed
	select {
	case e.c <- eventData{listener, err}:
		return true
	case <-e.ctx.Done():
		return false
	}
}

type priorityCtxKey struct{}

func WithPriority(ctx context.Context, priority int) context.Context {
	return context.WithValue(ctx, priorityCtxKey{}, priority)
}

type priorityLimiterBuilder struct {
	id string
}

func NewPriorityLimiter(delegate limits.Limiter) limits.Limiter {
	return &priorityLimiter{
		Limiter:     delegate,
		id:          "",
		backlogSize: 64,
		timeout:     time.Second,
		backlog:     util.NewPriorityDeque[*event](64),
	}
}

type priorityLimiter struct {
	limits.Limiter
	sync.Mutex
	id          string
	backlogSize int
	timeout     time.Duration
	backlog     util.Deque[*event]
}

func (p *priorityLimiter) Acquire(ctx context.Context) (limits.Listener, error) {
	listener, err := p.tryAcquire(ctx)
	if err == nil {
		return listener, nil
	}

	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	priority, _ := ctx.Value(priorityCtxKey{}).(int)
	ev := &event{
		ctx:      ctx,
		priority: priority,
		c:        make(chan eventData),
	}

	defer ev.cancel()

	p.Lock()
	outdated, ok := p.backlog.Offer(ev)
	p.Unlock()

	if !ok {
		return nil, errBacklogOverload //errBacklogOverload
	}

	if outdated != nil {
		outdated.signal(nil, errEvicted)
	}

	select {
	case data := <-ev.c:
		return data.listener, data.err
	case <-ctx.Done():
		return nil, errTimeout
	}
}

func (p *priorityLimiter) tryAcquire(ctx context.Context) (limits.Listener, error) {
	listener, err := p.Limiter.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context, result limits.Result) {
		listener(ctx, result)
		p.signal(ctx)
	}, nil
}

func (p *priorityLimiter) signal(context.Context) {
	p.Lock()
	candidate, ok := p.backlog.PeekFirst()
	timeout := 0

	for ; ok && candidate.Done(); candidate, ok = p.backlog.PeekFirst() {
		timeout++
		p.backlog.PollFirst()
	}

	if !ok {
		p.Unlock()
		return
	}

	listener, err := p.tryAcquire(candidate.ctx)
	if err == nil {
		p.backlog.PollFirst()
		p.Unlock()
		if !candidate.signal(listener, nil) {
			// it is possible that the wait request already timeout,
			// release the limit directly, or there will be a limit leak
			listener(candidate.ctx, limits.IGNORED)
		}
	} else {
		p.Unlock()
	}
}
