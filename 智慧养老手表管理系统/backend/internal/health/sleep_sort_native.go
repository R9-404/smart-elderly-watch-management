//go:build linux && amd64 && cgo && !purego

package health

/*
#cgo CXXFLAGS: -O3
#cgo LDFLAGS: -lstdc++
#include <stdint.h>
typedef struct { int64_t seconds; uint32_t nanos; uint32_t index; } suining_sleep_key;
void suining_sort_sleep(suining_sleep_key *rows, uint32_t count);
#cgo noescape suining_sort_sleep
#cgo nocallback suining_sort_sleep
*/
import "C"

func sortSleepNative(samples []*Sample) bool {
	if len(samples) < 2048 || uint64(len(samples)) > uint64(^uint32(0)) {
		return false
	}
	// Normal device uploads often arrive in time order. Avoid packing them.
	ordered := true
	previous := samples[0].Start
	if previous != previous.Round(0) {
		return false
	}
	for _, sample := range samples[1:] {
		start := sample.Start
		if start != start.Round(0) {
			return false
		}
		if start.Before(previous) {
			ordered = false
			break
		}
		previous = start
	}
	if ordered {
		return true
	}
	keys := make([]C.suining_sleep_key, len(samples))
	for i, sample := range samples {
		start := sample.Start
		seconds := start.Unix()
		// Preserve time.Time comparison for monotonic clocks and dates outside
		// the RFC 3339 year range used by persisted samples.
		if start != start.Round(0) || seconds < -62167219200 || seconds >= 253402300800 {
			return false
		}
		keys[i].seconds = C.int64_t(seconds)
		keys[i].nanos = C.uint32_t(start.Nanosecond())
		keys[i].index = C.uint32_t(i)
	}
	C.suining_sort_sleep(&keys[0], C.uint32_t(len(keys)))

	// Apply the sorted permutation in place. Replacing each visited index
	// with its own position marks the cycle as complete.
	for i := range keys {
		next := int(keys[i].index)
		if next == i {
			continue
		}
		saved := samples[i]
		at := i
		for next != i {
			samples[at] = samples[next]
			keys[at].index = C.uint32_t(at)
			at = next
			next = int(keys[at].index)
		}
		samples[at] = saved
		keys[at].index = C.uint32_t(at)
	}
	return true
}
