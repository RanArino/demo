package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Chunking metrics for observability
var (
	// ChunkingTotal tracks total chunking attempts by status (completed, failed)
	ChunkingTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ms_canvas_chunking_attempts_total",
			Help: "Total number of chunking attempts by status",
		},
		[]string{"status"}, // completed, failed
	)

	// ChunkingDuration tracks how long chunking operations take
	ChunkingDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "ms_canvas_chunking_duration_seconds",
			Help:    "Duration of chunking operations in seconds",
			Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
		},
	)

	// ChunkingRetries tracks automatic retry attempts
	ChunkingRetries = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "ms_canvas_chunking_retries_total",
			Help: "Total number of automatic chunking retry attempts",
		},
	)
)
