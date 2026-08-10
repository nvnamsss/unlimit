package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/voidforge-studios/unlimit/observe"
)

// ---------- mocks ----------

type mockSpan struct {
	tags  map[string]interface{}
	err   error
	ended bool
}

func newMockSpan() *mockSpan {
	return &mockSpan{tags: make(map[string]interface{})}
}

func (s *mockSpan) SetTag(key string, value interface{}) { s.tags[key] = value }
func (s *mockSpan) SetError(err error)                   { s.err = err }
func (s *mockSpan) End()                                 { s.ended = true }

// mockTraceSpan also implements observe.TraceSpan (exposes IDs).
type mockTraceSpan struct {
	mockSpan
	traceID string
	spanID  string
}

func (s *mockTraceSpan) TraceID() string { return s.traceID }
func (s *mockTraceSpan) SpanID() string  { return s.spanID }

type mockTracer struct {
	span observe.Span
}

func (t *mockTracer) StartSpan(ctx context.Context, name string) (context.Context, observe.Span) {
	return ctx, t.span
}

// ---------- helpers ----------

func setupRouter(handler gin.HandlerFunc, statusCode int) (*gin.Engine, *mockTracer) {
	gin.SetMode(gin.TestMode)
	tracer := &mockTracer{}
	r := gin.New()
	r.Use(handler)
	r.GET("/test", func(c *gin.Context) {
		c.Status(statusCode)
	})
	return r, tracer
}

func performRequest(r *gin.Engine) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
	return w
}

// ---------- tests ----------

func TestTracingMiddleware_NormalRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	span := newMockSpan()
	tracer := &mockTracer{span: span}

	r := gin.New()
	r.Use(TracingMiddleware(tracer))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if !span.ended {
		t.Error("expected span.End() to be called")
	}
	if span.err != nil {
		t.Errorf("expected no error, got %v", span.err)
	}
	if span.tags["http.method"] != http.MethodGet {
		t.Errorf("expected http.method=%q, got %v", http.MethodGet, span.tags["http.method"])
	}
	if span.tags["http.status_code"] != http.StatusOK {
		t.Errorf("expected http.status_code=200, got %v", span.tags["http.status_code"])
	}
	if span.tags["http.route"] == nil {
		t.Error("expected http.route to be set")
	}
}

func TestTracingMiddleware_5xxSetsError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	span := newMockSpan()
	tracer := &mockTracer{span: span}

	r := gin.New()
	r.Use(TracingMiddleware(tracer))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusInternalServerError)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if span.err == nil {
		t.Error("expected span.SetError to be called for 5xx response")
	}
	if span.tags["http.status_code"] != http.StatusInternalServerError {
		t.Errorf("expected http.status_code=500, got %v", span.tags["http.status_code"])
	}
	if !span.ended {
		t.Error("expected span.End() to be called")
	}
}

func TestTracingMiddleware_NilTracerIsNoop(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	r := gin.New()
	r.Use(TracingMiddleware(nil))
	r.GET("/test", func(c *gin.Context) {
		called = true
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if !called {
		t.Error("expected handler to be called when tracer is nil")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestTracingMiddleware_TraceSpanSetsHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	span := &mockTraceSpan{
		mockSpan: *newMockSpan(),
		traceID:  "trace-abc-123",
		spanID:   "span-def-456",
	}
	tracer := &mockTracer{span: span}

	r := gin.New()
	r.Use(TracingMiddleware(tracer))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if got := w.Header().Get("X-Trace-ID"); got != "trace-abc-123" {
		t.Errorf("expected X-Trace-ID=%q, got %q", "trace-abc-123", got)
	}
	if got := w.Header().Get("X-Span-ID"); got != "span-def-456" {
		t.Errorf("expected X-Span-ID=%q, got %q", "span-def-456", got)
	}
}
