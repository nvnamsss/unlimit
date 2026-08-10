package observe

import "context"

// MetricCollector defines the interface for collecting metrics
type MetricCollector interface {
	Counter(name string, value int64, tags ...string)
	Gauge(name string, value float64, tags ...string)
	Histogram(name string, value float64, tags ...string)
	Close() error
}

type Tracer interface {
	StartSpan(ctx context.Context, name string) (context.Context, Span)
}

// Span represents a single trace span
type Span interface {
	SetTag(key string, value interface{})
	SetError(err error)
	End()
}

// TraceSpan is an optional extension of Span that exposes trace and span IDs.
// Middleware can type-assert a Span to TraceSpan to propagate IDs via response headers.
type TraceSpan interface {
	Span
	TraceID() string
	SpanID() string
}

// Field represents a structured logging field
type Field struct {
	Key   string
	Value interface{}
}
