package limit

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/xtracker/limits"
	"github.com/xtracker/limits/limit/window"
	"github.com/xtracker/limits/util"
)

type windowedLimitBuilder struct {
	minWindowTime       time.Duration
	maxWindowTime       time.Duration
	minRttThreshold     time.Duration
	windowSize          int
	sampleWindowFactory func() window.SampleWindow
}

func NewWindowedLimitBuilder() *windowedLimitBuilder {
	return &windowedLimitBuilder{
		minWindowTime:       time.Second,
		maxWindowTime:       time.Second,
		minRttThreshold:     time.Microsecond * 100,
		windowSize:          10,
		sampleWindowFactory: window.NewAverageSampleWindow,
	}
}

func (w *windowedLimitBuilder) Build(delegate limits.Limit) limits.Limit {
	return &WindowedLimit{
		Limit:           delegate,
		minWindowTime:   w.minWindowTime,
		maxWindowTime:   w.maxWindowTime,
		minRttThreshold: w.minRttThreshold,
		windowSize:      w.windowSize,
		sample:          window.NewBufferedSampleWindow(w.sampleWindowFactory()),
	}
}

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

	if endTime.Before(wl.nextUpdateTime.Load().(time.Time)) ||
		!atomic.CompareAndSwapInt32(&wl.updating, 0, 1) {
		return
	}

	defer atomic.StoreInt32(&wl.updating, 0)
	if endTime.After(wl.nextUpdateTime.Load().(time.Time)) {
		nextUpdateTime := endTime.Add(util.Min(wl.maxWindowTime, util.Max(wl.minWindowTime, wl.sample.GetCandidateRttNanos())))
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
