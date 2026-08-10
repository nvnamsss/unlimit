package middlewares

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/voidforge-studios/unlimit/observe"
)

// TracingMiddleware returns a Gin middleware that starts a trace span for each
// request, propagates the context, and records HTTP tags on the span.
//
// If tracer is nil the middleware is a pass-through.
func TracingMiddleware(tracer observe.Tracer) gin.HandlerFunc {
	if tracer == nil {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		name := c.Request.Method + " " + route

		ctx, span := tracer.StartSpan(c.Request.Context(), name)
		defer span.End()

		c.Request = c.Request.WithContext(ctx)

		// Propagate trace/span IDs via response headers when the span supports it.
		if ts, ok := span.(observe.TraceSpan); ok {
			if id := ts.TraceID(); id != "" {
				c.Header("X-Trace-ID", id)
			}
			if id := ts.SpanID(); id != "" {
				c.Header("X-Span-ID", id)
			}
		}

		span.SetTag("http.method", c.Request.Method)
		span.SetTag("http.route", route)

		c.Next()

		status := c.Writer.Status()
		span.SetTag("http.status_code", status)

		if status >= 500 {
			span.SetError(fmt.Errorf("http error: status %d", status))
		}
	}
}
