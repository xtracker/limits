package limit

import (
	"context"
	"time"
)

type FixedLimit int

func (f FixedLimit) GetLimit() int {
	return int(f)
}

func (FixedLimit) OnSample(context.Context, time.Time, time.Duration, int, bool) {

}

func (FixedLimit) NotifyChange(func(int)) {

}

func (FixedLimit) String() string {
	return "fixed"
}
