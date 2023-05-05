package window

import (
	"runtime"
	"sync/atomic"
	"time"
	_ "unsafe"
)

// encode a sample using uint64
// rtt(48bit):inflight(15bit):drop(1bit)
type dataPoint uint64

func (dp dataPoint) sample() (time.Duration, int, bool) {
	return time.Duration(uint64(dp) >> 16), int(uint16(dp) >> 1), (uint64(dp)&0x01 == 1)
}

func makeDataPoint(rtt time.Duration, inflight int, didDrop bool) dataPoint {
	bits := uint64(rtt)<<16 | uint64(inflight<<1)
	if didDrop {
		bits |= 1
	}

	return dataPoint(bits)
}

// lock free ring buffer
// push & pop will not be access in same thread
type ring struct {
	dps        []dataPoint
	head, tail uint64
	size       uint64
}

func (r *ring) increment(cur uint64) uint64 {
	cur++
	if r.size&(r.size-1) == 0 {
		return (cur) & (r.size - 1)
	}

	return cur % r.size
}

func (r *ring) len() int {
	tail := atomic.LoadUint64(&r.tail)
	head := atomic.LoadUint64(&r.head)

	return int((tail + r.size - head) % (r.size))
}

func (r *ring) offer(dp dataPoint) bool {
	tail := atomic.LoadUint64(&r.tail)
	nextTail := r.increment(tail)
	if nextTail == atomic.LoadUint64(&r.head) {
		return false // full
	}

	r.dps[tail] = dp
	atomic.StoreUint64(&r.tail, nextTail)
	return true
}

func (r *ring) poll() (dataPoint, bool) {
	head := atomic.LoadUint64(&r.head)
	tail := atomic.LoadUint64(&r.tail)
	if head == tail {
		return 0, false // empty, direct return
	}

	nextHead := r.increment(head)
	dp := r.dps[head]
	atomic.StoreUint64(&r.tail, nextHead)
	return dp, true
}

// use iterator to fetch ring data
// poll will set mem barrier each element
type iterator interface {
	next() (dataPoint, bool)
	close()
}

type snapshotIterator struct {
	*ring
	target  uint64
	current uint64
}

func (si *snapshotIterator) next() (dataPoint, bool) {
	if si.current == si.target {
		return 0, false
	}

	dp := si.dps[si.current]
	si.current = si.increment(si.current)
	return dp, true
}

func (si *snapshotIterator) close() {
	atomic.StoreUint64(&si.head, si.current)
}

func (r *ring) snapshot() iterator {
	return &snapshotIterator{
		ring:    r,
		target:  atomic.LoadUint64(&r.tail),
		current: atomic.LoadUint64(&r.head),
	}
}

type DataPoints struct {
	ring
}

func NewBufferedSampleWindow(delegate SampleWindow) SampleWindow {
	return &bufferedSampleWindow{
		SampleWindow: delegate,
		dps:          make([]*DataPoints, runtime.GOMAXPROCS(0)),
	}
}

type bufferedSampleWindow struct {
	SampleWindow // stratrgy
	dps          []*DataPoints
}

func (bsw *bufferedSampleWindow) AddSample(rtt time.Duration, inflight int, dropped bool) {
	id := procPin()
	defer procUnPin()

	// protect if GOMAXPROCS were modified at runtime
	if id < len(bsw.dps) {
		bsw.dps[id].offer(makeDataPoint(rtt, inflight, dropped))
	}
}

func (bsw *bufferedSampleWindow) SnapShot() SampleWindow {
	bsw.SampleWindow.Reset()
	bsw.flush()
	return bsw.SampleWindow
}

func (bsw *bufferedSampleWindow) flush() {
	for _, s := range bsw.dps {
		bsw.flushSlot(s)
	}
}

func (bsw *bufferedSampleWindow) flushSlot(dps *DataPoints) {
	iter := dps.snapshot()
	defer iter.close()
	for dp, ok := iter.next(); ok; dp, ok = iter.next() {
		rtt, inflight, didDrop := dp.sample()
		bsw.SampleWindow.AddSample(rtt, inflight, didDrop)
	}
}

func (bsw *bufferedSampleWindow) GetCandidateRttNanos() time.Duration {
	return bsw.SampleWindow.GetCandidateRttNanos()
}

func (bsw *bufferedSampleWindow) GetTrackedRttNanos() time.Duration {
	return bsw.GetTrackedRttNanos()
}

func (bsw *bufferedSampleWindow) GetMaxInFlight() int {
	return bsw.SampleWindow.GetMaxInFlight()
}

func (bsw *bufferedSampleWindow) GetSampleCount() (int, int) {
	sc := 0
	for _, dps := range bsw.dps {
		sc += dps.len()
	}

	return sc, 0
}

func (bsw *bufferedSampleWindow) DidDrop() bool {
	return bsw.SampleWindow.DidDrop()
}

func (bsw *bufferedSampleWindow) Reset() {
}

//go:linkname procPin runtime.procPin
func procPin() int

//go:linkname procUnPin runtime.procUnpin
func procUnPin()
