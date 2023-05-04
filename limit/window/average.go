package window

import (
	"time"

	"github.com/xtracker/limits/util"
)

type AverageSampleWindow struct {
	base
	sumRtt time.Duration
}

func (a *AverageSampleWindow) AddSample(rtt time.Duration, inflight int, dropped bool) {
	if dropped {
		a.dropped++
	} else {
		a.sumRtt += rtt
	}

	a.sampleCount++
	a.maxInFlight = util.Max(a.maxInFlight, inflight)
}

func (a *AverageSampleWindow) GetTrackedRttNanos() time.Duration {
	if a.sampleCount == a.dropped {
		return 0
	}

	return a.sumRtt / time.Duration(a.sampleCount-a.dropped)
}
