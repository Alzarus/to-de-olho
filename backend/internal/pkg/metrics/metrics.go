package metrics

import (
	"sync/atomic"
)

// Simple in-memory counters for small project-level metrics used in tests and logs.
// This is intentionally minimal: if Prometheus is enabled elsewhere, it can be wired
// to these counters or replaced by a fuller implementation.

var schedulerSkipCounters = map[string]*int64{}

// IncSchedulerSkip increments the skip counter for a scheduler tipo.
func IncSchedulerSkip(tipo string) {
	key := tipo
	p, ok := schedulerSkipCounters[key]
	if !ok {
		var v int64 = 0
		schedulerSkipCounters[key] = &v
		p = &v
	}
	atomic.AddInt64(p, 1)
}

// GetSchedulerSkips returns the current skip count for a tipo.
func GetSchedulerSkips(tipo string) int64 {
	if p, ok := schedulerSkipCounters[tipo]; ok {
		return atomic.LoadInt64(p)
	}
	return 0
}
