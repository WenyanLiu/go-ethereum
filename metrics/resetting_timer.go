package metrics

import (
	"math"
	"sort"
	"sync"
	"time"
	"github.com/ethereum/go-ethereum/log"
)

// Initial slice capacity for the values stored in a ResettingTimer
const InitialResettingTimerSliceCap = 10

// ResettingTimer is used for storing aggregated values for timers, which are reset on every flush interval.
type ResettingTimer interface {
	Values() []int64
	Snapshot() ResettingTimer
	Percentiles([]float64) []int64
	Mean() float64
	Time(func())
	Update(time.Duration)
	UpdateSince(time.Time)
}

// GetOrRegisterResettingTimer returns an existing ResettingTimer or constructs and registers a
// new StandardResettingTimer.
func GetOrRegisterResettingTimer(name string, r Registry) ResettingTimer {
	log.DebugLog()
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, NewResettingTimer).(ResettingTimer)
}

// NewRegisteredResettingTimer constructs and registers a new StandardResettingTimer.
func NewRegisteredResettingTimer(name string, r Registry) ResettingTimer {
	log.DebugLog()
	c := NewResettingTimer()
	if nil == r {
		r = DefaultRegistry
	}
	r.Register(name, c)
	return c
}

// NewResettingTimer constructs a new StandardResettingTimer
func NewResettingTimer() ResettingTimer {
	log.DebugLog()
	if !Enabled {
		return NilResettingTimer{}
	}
	return &StandardResettingTimer{
		values: make([]int64, 0, InitialResettingTimerSliceCap),
	}
}

// NilResettingTimer is a no-op ResettingTimer.
type NilResettingTimer struct {
}

// Values is a no-op.
func (NilResettingTimer) Values() []int64 { log.DebugLog()
											  return nil }

// Snapshot is a no-op.
func (NilResettingTimer) Snapshot() ResettingTimer { log.DebugLog()
													   return NilResettingTimer{} }

// Time is a no-op.
func (NilResettingTimer) Time(func()) { log.DebugLog() }

// Update is a no-op.
func (NilResettingTimer) Update(time.Duration) { log.DebugLog() }

// Percentiles panics.
func (NilResettingTimer) Percentiles([]float64) []int64 {
	log.DebugLog()
	panic("Percentiles called on a NilResettingTimer")
}

// Mean panics.
func (NilResettingTimer) Mean() float64 {
	log.DebugLog()
	panic("Mean called on a NilResettingTimer")
}

// UpdateSince is a no-op.
func (NilResettingTimer) UpdateSince(time.Time) { log.DebugLog() }

// StandardResettingTimer is the standard implementation of a ResettingTimer.
// and Meter.
type StandardResettingTimer struct {
	values []int64
	mutex  sync.Mutex
}

// Values returns a slice with all measurements.
func (t *StandardResettingTimer) Values() []int64 {
	log.DebugLog()
	return t.values
}

// Snapshot resets the timer and returns a read-only copy of its contents.
func (t *StandardResettingTimer) Snapshot() ResettingTimer {
	log.DebugLog()
	t.mutex.Lock()
	defer t.mutex.Unlock()
	currentValues := t.values
	t.values = make([]int64, 0, InitialResettingTimerSliceCap)

	return &ResettingTimerSnapshot{
		values: currentValues,
	}
}

// Percentiles panics.
func (t *StandardResettingTimer) Percentiles([]float64) []int64 {
	log.DebugLog()
	panic("Percentiles called on a StandardResettingTimer")
}

// Mean panics.
func (t *StandardResettingTimer) Mean() float64 {
	log.DebugLog()
	panic("Mean called on a StandardResettingTimer")
}

// Record the duration of the execution of the given function.
func (t *StandardResettingTimer) Time(f func()) {
	log.DebugLog()
	ts := time.Now()
	f()
	t.Update(time.Since(ts))
}

// Record the duration of an event.
func (t *StandardResettingTimer) Update(d time.Duration) {
	log.DebugLog()
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.values = append(t.values, int64(d))
}

// Record the duration of an event that started at a time and ends now.
func (t *StandardResettingTimer) UpdateSince(ts time.Time) {
	log.DebugLog()
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.values = append(t.values, int64(time.Since(ts)))
}

// ResettingTimerSnapshot is a point-in-time copy of another ResettingTimer.
type ResettingTimerSnapshot struct {
	values              []int64
	mean                float64
	thresholdBoundaries []int64
	calculated          bool
}

// Snapshot returns the snapshot.
func (t *ResettingTimerSnapshot) Snapshot() ResettingTimer { log.DebugLog()
															   return t }

// Time panics.
func (*ResettingTimerSnapshot) Time(func()) {
	log.DebugLog()
	panic("Time called on a ResettingTimerSnapshot")
}

// Update panics.
func (*ResettingTimerSnapshot) Update(time.Duration) {
	log.DebugLog()
	panic("Update called on a ResettingTimerSnapshot")
}

// UpdateSince panics.
func (*ResettingTimerSnapshot) UpdateSince(time.Time) {
	log.DebugLog()
	panic("UpdateSince called on a ResettingTimerSnapshot")
}

// Values returns all values from snapshot.
func (t *ResettingTimerSnapshot) Values() []int64 {
	log.DebugLog()
	return t.values
}

// Percentiles returns the boundaries for the input percentiles.
func (t *ResettingTimerSnapshot) Percentiles(percentiles []float64) []int64 {
	log.DebugLog()
	t.calc(percentiles)

	return t.thresholdBoundaries
}

// Mean returns the mean of the snapshotted values
func (t *ResettingTimerSnapshot) Mean() float64 {
	log.DebugLog()
	if !t.calculated {
		t.calc([]float64{})
	}

	return t.mean
}

func (t *ResettingTimerSnapshot) calc(percentiles []float64) {
	log.DebugLog()
	sort.Sort(Int64Slice(t.values))

	count := len(t.values)
	if count > 0 {
		min := t.values[0]
		max := t.values[count-1]

		cumulativeValues := make([]int64, count)
		cumulativeValues[0] = min
		for i := 1; i < count; i++ {
			cumulativeValues[i] = t.values[i] + cumulativeValues[i-1]
		}

		t.thresholdBoundaries = make([]int64, len(percentiles))

		thresholdBoundary := max

		for i, pct := range percentiles {
			if count > 1 {
				var abs float64
				if pct >= 0 {
					abs = pct
				} else {
					abs = 100 + pct
				}
				// poor man's math.Round(x):
				// math.Floor(x + 0.5)
				indexOfPerc := int(math.Floor(((abs / 100.0) * float64(count)) + 0.5))
				if pct >= 0 {
					indexOfPerc -= 1 // index offset=0
				}
				thresholdBoundary = t.values[indexOfPerc]
			}

			t.thresholdBoundaries[i] = thresholdBoundary
		}

		sum := cumulativeValues[count-1]
		t.mean = float64(sum) / float64(count)
	} else {
		t.thresholdBoundaries = make([]int64, len(percentiles))
		t.mean = 0
	}

	t.calculated = true
}

// Int64Slice attaches the methods of sort.Interface to []int64, sorting in increasing order.
type Int64Slice []int64

func (s Int64Slice) Len() int           { log.DebugLog()
											return len(s) }
func (s Int64Slice) Less(i, j int) bool { log.DebugLog()
											return s[i] < s[j] }
func (s Int64Slice) Swap(i, j int)      { log.DebugLog()
											s[i], s[j] = s[j], s[i] }
