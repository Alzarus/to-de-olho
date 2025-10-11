package metrics

import (
	"fmt"
	"strconv"
	"sync"
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

var (
	camaraChunkSuccess sync.Map
	camaraChunkFailure sync.Map
)

func incrementChunkCounter(store *sync.Map, window int) int64 {
	key := fmt.Sprint(window)
	val, _ := store.LoadOrStore(key, new(int64))
	ptr := val.(*int64)
	return atomic.AddInt64(ptr, 1)
}

// IncCamaraChunkSuccess registra o sucesso de uma janela na coleta da API da Câmara.
func IncCamaraChunkSuccess(windowDays int) int64 {
	return incrementChunkCounter(&camaraChunkSuccess, windowDays)
}

// IncCamaraChunkFailure registra falha (ex.: HTTP 504) em uma janela específica.
func IncCamaraChunkFailure(windowDays int) int64 {
	return incrementChunkCounter(&camaraChunkFailure, windowDays)
}

// SnapshotCamaraChunkMetrics retorna um mapa janela→{success,failure} para observabilidade simples.
func SnapshotCamaraChunkMetrics() map[int]struct {
	Success int64
	Failure int64
} {
	result := make(map[int]struct {
		Success int64
		Failure int64
	})

	camaraChunkSuccess.Range(func(key, value any) bool {
		window, err := strconv.Atoi(key.(string))
		if err != nil {
			return true
		}
		metric := result[window]
		metric.Success = atomic.LoadInt64(value.(*int64))
		result[window] = metric
		return true
	})

	camaraChunkFailure.Range(func(key, value any) bool {
		window, err := strconv.Atoi(key.(string))
		if err != nil {
			return true
		}
		metric := result[window]
		metric.Failure = atomic.LoadInt64(value.(*int64))
		result[window] = metric
		return true
	})

	return result
}
