package window

import (
	"time"

	"github.com/xtracker/limits/util"
)

type base struct {
	minRtt      time.Duration
	maxInFlight int
	sampleCount int
	dropped     int
	rate        int
}

func (w *base) AddSample(rtt time.Duration, inflight int, dropped bool) {
	if !dropped {
		w.minRtt = util.Min(rtt, w.minRtt)
	} else {
		w.dropped++
	}

	w.sampleCount++
}

func (w *base) DidDrop() bool {
	return w.dropped*100 > w.sampleCount*w.rate
}

func (w *base) GetMaxInFlight() int {
	return w.maxInFlight
}

func (w *base) GetCandidateRttNanos() time.Duration {
	return w.minRtt
}

func (w *base) GetSampleCount() (int, int) {
	return w.sampleCount, w.dropped
}

func (w *base) Reset() {
	w.minRtt = time.Hour
	w.maxInFlight = 0
	w.sampleCount = 0
	w.dropped = 0
	w.rate = 0
}
