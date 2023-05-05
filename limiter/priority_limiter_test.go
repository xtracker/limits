package limiter

import (
	"context"
	"testing"
	"time"

	"github.com/xtracker/limits/limit"
)

func TestContextValue(t *testing.T) {
	priority, _ := context.Background().Value("key").(int)
	t.Fatalf("%d", priority)
}

func TestPriorityLimiter(t *testing.T) {

	limitAlgorithm := limit.FixedLimit(1)
	pl := NewPriorityLimiterBuilder(NewSimpleLimiter("", limitAlgorithm)).Build()

	ctx := context.Background()
	start := time.Now()
	pl.Acquire(ctx)
	elapse := time.Since(start)
	if elapse > time.Millisecond {
		t.Fail()
	}

	start = time.Now()
	pl.Acquire(ctx)
	elapse = time.Since(start)
	if elapse > time.Second / 2 {
		t.Fail()
	}
}
