package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
)

func resetStats() {
	mu.Lock()
	defer mu.Unlock()
	stats = make(map[string]*endpointStats)
}

func TestMiddlewareRecordsRequests(t *testing.T) {
	resetStats()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/secret/abc", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/secret/:hash")

	handler := Middleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	handler(c)
	handler(c)

	mu.RLock()
	s, ok := stats["GET /secret/:hash"]
	mu.RUnlock()

	if !ok {
		t.Fatal("expected stats entry for GET /secret/:hash")
	}
	if s.count != 2 {
		t.Fatalf("expected 2 requests, got %d", s.count)
	}
	if s.totalNs <= 0 {
		t.Fatal("expected positive total latency")
	}
}

func TestMiddlewareTracksMultipleEndpoints(t *testing.T) {
	resetStats()

	e := echo.New()
	callHandler := func(method, target, path string) {
		req := httptest.NewRequest(method, target, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath(path)
		Middleware()(func(c echo.Context) error { return nil })(c)
	}

	callHandler(http.MethodPost, "/secret", "/secret")
	callHandler(http.MethodGet, "/secret/abc", "/secret/:hash")
	callHandler(http.MethodGet, "/secret/xyz", "/secret/:hash")

	mu.RLock()
	defer mu.RUnlock()

	if stats["POST /secret"].count != 1 {
		t.Fatalf("expected 1 POST /secret request, got %d", stats["POST /secret"].count)
	}
	if stats["GET /secret/:hash"].count != 2 {
		t.Fatalf("expected 2 GET /secret/:hash requests, got %d", stats["GET /secret/:hash"].count)
	}
}

func TestHandlerReturnsJSON(t *testing.T) {
	resetStats()

	mu.Lock()
	s := &endpointStats{count: 4, totalNs: 8_000_000} // 8ms total → 2ms avg
	stats["POST /secret"] = s
	mu.Unlock()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := Handler()(c); err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	var result map[string]struct {
		Requests     int64   `json:"requests"`
		AvgLatencyMs float64 `json:"avg_latency_ms"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	entry, ok := result["POST /secret"]
	if !ok {
		t.Fatal("expected POST /secret in metrics response")
	}
	if entry.Requests != 4 {
		t.Fatalf("expected 4 requests, got %d", entry.Requests)
	}
	if entry.AvgLatencyMs != 2.0 {
		t.Fatalf("expected avg latency 2.0ms, got %f", entry.AvgLatencyMs)
	}
}

func TestHandlerEmptyStats(t *testing.T) {
	resetStats()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := Handler()(c); err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
