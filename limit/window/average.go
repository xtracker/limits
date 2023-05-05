package window

import (
	"time"
)

func NewAverageSampleWindow() SampleWindow {
	return &AverageSampleWindow{
		base: base{
			minRtt: time.Hour,
		},
	}
}

type AverageSampleWindow struct {
	base
	sumRtt time.Duration
}

func (a *AverageSampleWindow) AddSample(rtt time.Duration, inflight int, dropped bool) {
	a.base.AddSample(rtt, inflight, dropped)

	if !dropped {
		a.sumRtt += rtt
	}
}

func (a *AverageSampleWindow) GetTrackedRttNanos() time.Duration {
	if a.sampleCount == a.dropped {
		return 0
	}

	return a.sumRtt / time.Duration(a.sampleCount-a.dropped)
}

func (a *AverageSampleWindow) SnapShot() SampleWindow {
	return a
}
