package limit

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/xtracker/limits"
	"github.com/xtracker/limits/limit/window"
)

type WindowedLimit struct {
	limits.Limit
	nextUpdateTime  atomic.Value //time.Time
	minWindowTime   time.Duration
	maxWindowTime   time.Duration
	minRttThreshold time.Duration
	sample          window.SampleWindow
	windowSize      int
	updating        int32
}

func (wl *WindowedLimit) OnSample(ctx context.Context, startTime time.Time, rtt time.Duration, inflight int, dropped bool) {
	if rtt < wl.minRttThreshold {
		return
	}

	wl.sample.AddSample(rtt, inflight, dropped)

	endTime := startTime.Add(rtt)

	if !endTime.After(wl.nextUpdateTime.Load().(time.Time)) ||
		!atomic.CompareAndSwapInt32(&wl.updating, 0, 1) {
		return
	}

	defer atomic.StoreInt32(&wl.updating, 0)
	if endTime.After(wl.nextUpdateTime.Load().(time.Time)) {
		nextUpdateTime := endTime.Add(time.Second)
		wl.nextUpdateTime.Store(nextUpdateTime)
		if wl.isWindowReady(wl.sample) {
			sample := wl.sample.SnapShot()
			wl.Limit.OnSample(ctx, startTime, sample.GetTrackedRttNanos(),
				sample.GetMaxInFlight(), sample.DidDrop())
			wl.sample.Reset()
		}

	}
}

func (wl *WindowedLimit) isWindowReady(sw window.SampleWindow) bool {
	total, _ := sw.GetSampleCount()
	return total >= wl.windowSize
}
