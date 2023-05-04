package window

import "time"

type SampleWindow interface {
	AddSample(rtt time.Duration, inflight int, didDrop bool)

	SnapShot() SampleWindow

	GetCandidateRttNanos() time.Duration

	GetTrackedRttNanos() time.Duration

	GetMaxInFlight() int

	GetSampleCount() (int, int)

	DidDrop() bool

	Reset()
}
