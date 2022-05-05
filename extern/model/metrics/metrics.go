package metrics

import (
	"context"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

var (
	Name, _  = tag.NewKey("name")  // name of running instance of visor
	Table, _ = tag.NewKey("table") // name of table data is persisted for
	API, _   = tag.NewKey("api")   // name of method on lotus api
)

var (
	PersistDuration = stats.Float64("persist_duration_ms", "Duration of a models persist operation", stats.UnitMilliseconds)
	PersistModel    = stats.Int64("persist_model", "Number of models persisted", stats.UnitDimensionless)
)

// SinceInMilliseconds returns the duration of time since the provide time as a float64.
func SinceInMilliseconds(startTime time.Time) float64 {
	return float64(time.Since(startTime).Nanoseconds()) / 1e6
}

// Timer is a function stopwatch, calling it starts the timer,
// calling the returned function will record the duration.
func Timer(ctx context.Context, m *stats.Float64Measure) func() {
	start := time.Now()
	return func() {
		stats.Record(ctx, m.M(SinceInMilliseconds(start)))
	}
}

// RecordCount is a convenience function that increments a counter by a count.
func RecordCount(ctx context.Context, m *stats.Int64Measure, count int) {
	stats.Record(ctx, m.M(int64(count)))
}
