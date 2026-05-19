package metrics

import (
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/labstack/echo"
)

type endpointStats struct {
	count   int64
	totalNs int64
}

var (
	mu    sync.RWMutex
	stats = make(map[string]*endpointStats)
)

func record(key string, elapsed time.Duration) {
	mu.Lock()
	s, ok := stats[key]
	if !ok {
		s = &endpointStats{}
		stats[key] = s
	}
	mu.Unlock()

	atomic.AddInt64(&s.count, 1)
	atomic.AddInt64(&s.totalNs, int64(elapsed))
}

// Middleware records per-request count and latency for every endpoint.
func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			record(c.Request().Method+" "+c.Path(), time.Since(start))
			return err
		}
	}
}

// Handler returns aggregated request counts and average latency per endpoint.
func Handler() echo.HandlerFunc {
	return func(c echo.Context) error {
		mu.RLock()
		defer mu.RUnlock()

		type stat struct {
			Requests     int64   `json:"requests"`
			AvgLatencyMs float64 `json:"avg_latency_ms"`
		}

		result := make(map[string]stat)
		for k, s := range stats {
			count := atomic.LoadInt64(&s.count)
			totalNs := atomic.LoadInt64(&s.totalNs)
			avg := float64(0)
			if count > 0 {
				avg = float64(totalNs) / float64(count) / 1e6
			}
			result[k] = stat{Requests: count, AvgLatencyMs: avg}
		}

		return c.JSON(http.StatusOK, result)
	}
}
