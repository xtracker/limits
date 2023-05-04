package limit

import (
	"context"
	"math"
	"time"

	"github.com/xtracker/limits/limit/measurement"
)

func NewGradient2Limit() *Gradient2Limit {
	return &Gradient2Limit{}
}

type Gradient2Limit struct {
	baseLimit

	/**
	     * Estimated concurrency limit based on our algorithm
		 * Need volatile
	*/
	estimatedLimit float64

	/**
	 * Tracks a measurement of the short time, and more volatile, RTT meant to represent the current system latency
	 */
	lastRtt time.Duration

	/**
	 * Tracks a measurement of the long term, less volatile, RTT meant to represent the baseline latency.  When the system
	 * is under load gl number is expect to trend higher.
	 */
	longRtt measurement.Measurement

	/**
	 * Maximum allowed limit providing an upper bound failsafe
	 */
	initLimit, minLimit, maxLimit float64

	queueSize func(int) float64

	smoothing float64

	tolerance float64
}

func (gl *Gradient2Limit) OnSample(ctx context.Context, startTime time.Time, rtt, sumRtt time.Duration, inflight int, sample int, didDrop bool) {
	defer gl.setLimit(int(gl.estimatedLimit))
	queueSize := gl.queueSize(int(gl.estimatedLimit))
	appLimited := inflight < int(gl.estimatedLimit/2.0)

	longRtt := gl.longRtt.Add(measurement.Int64Number(rtt)).Float64()
	gl.lastRtt = rtt
	shortRtt := float64(rtt)

	// If the long RTT is substantially larger than the short RTT then reduce the long RTT measurement.
	// gl can happen when latency returns to normal after a prolonged prior of excessive load.  Reducing the
	// long RTT without waiting for the exponential smoothing helps bring the system back to steady state.
	if longRtt/shortRtt > 2 {
		/*	logs.Infof("[%v] recovered, inflight:%v, limit:%.4f, shortRtt:%v, longRtt:%v, didDrop:%v",
			gl.id, inflight, gl.estimatedLimit, rtt, time.Duration(gl.longRtt.Get().Int64()), didDrop)*/
		gl.longRtt.Update(func(current measurement.Number) measurement.Number {
			return measurement.Float64Number(current.Float64() * 0.95)
		})
	}

	// Don't grow the limit if we are app limited
	if appLimited {
		return
	}

	// Rtt could be higher than rtt_noload because of smoothing rtt noload updates
	// so set to 1.0 to indicate no queuing.  Otherwise calculate the slope and don't
	// allow it to be reduced by more than half to avoid aggressive load-shedding due to
	// outliers.
	gradient := math.Max(0.5, math.Min(1, gl.tolerance*longRtt/shortRtt))

	newLimit := gl.estimatedLimit*gradient + queueSize
	newLimit = gl.estimatedLimit*(1-gl.smoothing) + newLimit*gl.smoothing
	newLimit = math.Max(gl.minLimit, math.Min(gl.maxLimit, newLimit))

	gl.estimatedLimit = newLimit
}
