package observe

import (
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusCollector implements MetricCollector using Prometheus.
// It provides thread-safe metric collection with support for counters, gauges, and histograms.
// All metrics are registered in a custom Prometheus registry and can be exposed via HTTP endpoints.
type PrometheusCollector struct {
	registry   *prometheus.Registry
	counters   map[string]*prometheus.CounterVec
	gauges     map[string]*prometheus.GaugeVec
	histograms map[string]*prometheus.HistogramVec
	mu         sync.RWMutex
}

// NewPrometheusCollector creates a new Prometheus metric collector with a fresh registry.
// It initializes empty maps for counters, gauges, and histograms that will be populated
// on-demand when metrics are first recorded.
// Example usage:
//
//	collector := NewPrometheusCollector()
//	collector.Counter("requests_total", 1, "method:GET", "status:200")
//
// This function is useful for creating an isolated metric collector that doesn't
// interfere with the default Prometheus registry.
func NewPrometheusCollector() *PrometheusCollector {
	return &PrometheusCollector{
		registry:   prometheus.NewRegistry(),
		counters:   make(map[string]*prometheus.CounterVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
		histograms: make(map[string]*prometheus.HistogramVec),
	}
}

// NewPrometheusCollectorWithRegistry creates a collector with a custom Prometheus registry.
// This allows you to use an existing registry or share a registry across multiple collectors.
// Example usage:
//
//	reg := prometheus.NewRegistry()
//	collector := NewPrometheusCollectorWithRegistry(reg)
//	collector.Gauge("memory_usage_bytes", 1024000, "service:api")
//
// This function is useful when you need to integrate with existing Prometheus infrastructure
// or when you want multiple collectors to share the same registry.
func NewPrometheusCollectorWithRegistry(registry *prometheus.Registry) *PrometheusCollector {
	return &PrometheusCollector{
		registry:   registry,
		counters:   make(map[string]*prometheus.CounterVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
		histograms: make(map[string]*prometheus.HistogramVec),
	}
}

// Counter increments a counter metric by the specified value.
// Counters are cumulative metrics that can only increase (or reset to zero on restart).
// Tags are optional and should be in the format "key:value" or "key=value".
// Example usage:
//
//	collector.Counter("http_requests_total", 1, "method:GET", "endpoint:/api/users")
//	collector.Counter("errors_total", 1, "type:validation", "severity:high")
//
// This function is useful for tracking cumulative values like request counts,
// error counts, or any metric that only increases over time.
// The metric is created automatically on first use and is thread-safe.
func (p *PrometheusCollector) Counter(name string, value int64, tags ...string) {
	counter := p.getOrCreateCounter(name, tags)
	if counter != nil {
		labels := p.parseLabels(tags)
		counter.With(labels).Add(float64(value))
	}
}

// Gauge sets a gauge metric to the specified value.
// Gauges represent values that can increase or decrease arbitrarily.
// Tags are optional and should be in the format "key:value" or "key=value".
// Example usage:
//
//	collector.Gauge("memory_usage_bytes", 1024000, "service:api", "host:server1")
//	collector.Gauge("active_connections", 42, "pool:database")
//
// This function is useful for tracking current values like memory usage,
// active connections, queue length, or any metric that can go up and down.
// The metric is created automatically on first use and is thread-safe.
func (p *PrometheusCollector) Gauge(name string, value float64, tags ...string) {
	gauge := p.getOrCreateGauge(name, tags)
	if gauge != nil {
		labels := p.parseLabels(tags)
		gauge.With(labels).Set(value)
	}
}

// Histogram records a histogram observation for the specified value.
// Histograms track the distribution of values and automatically calculate quantiles.
// Tags are optional and should be in the format "key:value" or "key=value".
// Example usage:
//
//	collector.Histogram("request_duration_seconds", 0.245, "method:GET", "endpoint:/api/users")
//	collector.Histogram("response_size_bytes", 4096, "content_type:json")
//
// This function is useful for tracking distributions like request durations,
// response sizes, or any metric where you need to understand the statistical
// distribution of values. Uses Prometheus default buckets for quantile calculation.
// The metric is created automatically on first use and is thread-safe.
func (p *PrometheusCollector) Histogram(name string, value float64, tags ...string) {
	histogram := p.getOrCreateHistogram(name, tags)
	if histogram != nil {
		labels := p.parseLabels(tags)
		histogram.With(labels).Observe(value)
	}
}

// Close implements the Close method required by the MetricCollector interface.
// For Prometheus, this is a no-op as the collector doesn't maintain any connections
// or resources that need explicit cleanup.
// Example usage:
//
//	defer collector.Close()
//
// This function always returns nil and is safe to call multiple times.
func (p *PrometheusCollector) Close() error {
	return nil
}

// Registry returns the underlying Prometheus registry used by this collector.
// This is useful when you need direct access to the registry for advanced use cases
// like custom metric gathering or integration with other Prometheus components.
// Example usage:
//
//	reg := collector.Registry()
//	metrics, _ := reg.Gather()
//
// This function is useful for accessing the raw Prometheus registry when you need
// to perform operations not exposed by the MetricCollector interface.
func (p *PrometheusCollector) Registry() *prometheus.Registry {
	return p.registry
}

// getOrCreateCounter retrieves an existing counter metric or creates a new one if it doesn't exist.
// This method uses double-checked locking for thread-safe lazy initialization of metrics.
// It extracts label names from tags and registers the counter with the collector's registry.
// The created counter includes all label names found in the tags for future label combinations.
// This is an internal method used by the Counter method to ensure metrics are created on-demand.
func (p *PrometheusCollector) getOrCreateCounter(name string, tags []string) *prometheus.CounterVec {
	p.mu.RLock()
	counter, exists := p.counters[name]
	p.mu.RUnlock()

	if exists {
		return counter
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock
	if counter, exists := p.counters[name]; exists {
		return counter
	}

	labelNames := p.extractLabelNames(tags)
	counter = promauto.With(p.registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: sanitizeMetricName(name),
			Help: fmt.Sprintf("Counter metric for %s", name),
		},
		labelNames,
	)

	p.counters[name] = counter
	return counter
}

// getOrCreateGauge retrieves an existing gauge metric or creates a new one if it doesn't exist.
// This method uses double-checked locking for thread-safe lazy initialization of metrics.
// It extracts label names from tags and registers the gauge with the collector's registry.
// The created gauge includes all label names found in the tags for future label combinations.
// This is an internal method used by the Gauge method to ensure metrics are created on-demand.
func (p *PrometheusCollector) getOrCreateGauge(name string, tags []string) *prometheus.GaugeVec {
	p.mu.RLock()
	gauge, exists := p.gauges[name]
	p.mu.RUnlock()

	if exists {
		return gauge
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock
	if gauge, exists := p.gauges[name]; exists {
		return gauge
	}

	labelNames := p.extractLabelNames(tags)
	gauge = promauto.With(p.registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Name: sanitizeMetricName(name),
			Help: fmt.Sprintf("Gauge metric for %s", name),
		},
		labelNames,
	)

	p.gauges[name] = gauge
	return gauge
}

// getOrCreateHistogram retrieves an existing histogram metric or creates a new one if it doesn't exist.
// This method uses double-checked locking for thread-safe lazy initialization of metrics.
// It extracts label names from tags and registers the histogram with the collector's registry.
// The created histogram uses Prometheus default buckets and includes all label names found in the tags.
// This is an internal method used by the Histogram method to ensure metrics are created on-demand.
func (p *PrometheusCollector) getOrCreateHistogram(name string, tags []string) *prometheus.HistogramVec {
	p.mu.RLock()
	histogram, exists := p.histograms[name]
	p.mu.RUnlock()

	if exists {
		return histogram
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock
	if histogram, exists := p.histograms[name]; exists {
		return histogram
	}

	labelNames := p.extractLabelNames(tags)
	histogram = promauto.With(p.registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    sanitizeMetricName(name),
			Help:    fmt.Sprintf("Histogram metric for %s", name),
			Buckets: prometheus.DefBuckets,
		},
		labelNames,
	)

	p.histograms[name] = histogram
	return histogram
}

// parseLabels converts tag strings to Prometheus labels map.
// It supports two tag formats: "key:value" and "key=value".
// Tags that don't match either format are ignored.
// Example usage:
//
//	tags := []string{"method:GET", "status=200", "endpoint:/api/users"}
//	labels := collector.parseLabels(tags)
//	// Returns: prometheus.Labels{"method": "GET", "status": "200", "endpoint": "/api/users"}
//
// This is an internal method used to convert the flexible tag format into
// the structured label format required by Prometheus.
func (p *PrometheusCollector) parseLabels(tags []string) prometheus.Labels {
	labels := make(prometheus.Labels)

	for _, tag := range tags {
		// Try colon separator first
		if len(tag) > 0 {
			key, value := "", ""
			for i, ch := range tag {
				if ch == ':' || ch == '=' {
					key = tag[:i]
					if i+1 < len(tag) {
						value = tag[i+1:]
					}
					break
				}
			}

			if key != "" {
				labels[key] = value
			}
		}
	}

	return labels
}

// extractLabelNames extracts unique label names from tags for metric registration.
// It parses tags in the format "key:value" or "key=value" and returns only the keys.
// Duplicate label names are filtered out to ensure each label is registered only once.
// Example usage:
//
//	tags := []string{"method:GET", "status=200", "method:POST"}
//	labels := collector.extractLabelNames(tags)
//	// Returns: ["method", "status"]
//
// This is an internal method used during metric registration to determine which
// label dimensions the metric should support.
func (p *PrometheusCollector) extractLabelNames(tags []string) []string {
	labelNames := make([]string, 0, len(tags))
	seen := make(map[string]bool)

	for _, tag := range tags {
		for i, ch := range tag {
			if ch == ':' || ch == '=' {
				key := tag[:i]
				if key != "" && !seen[key] {
					labelNames = append(labelNames, key)
					seen[key] = true
				}
				break
			}
		}
	}

	return labelNames
}

// sanitizeMetricName replaces invalid characters in metric names with underscores.
// Prometheus metric names must match the regex [a-zA-Z_:][a-zA-Z0-9_:]*.
// This function converts dots, hyphens, and spaces to underscores, and removes other invalid characters.
// Example usage:
//
//	name := sanitizeMetricName("http.request-count")
//	// Returns: "http_request_count"
//
// This function is useful for automatically converting metric names from other
// monitoring systems or human-readable formats into Prometheus-compliant names.
func sanitizeMetricName(name string) string {
	result := make([]byte, 0, len(name))

	for i := 0; i < len(name); i++ {
		ch := name[i]
		if (ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '_' {
			result = append(result, ch)
		} else if ch == '.' || ch == '-' || ch == ' ' {
			result = append(result, '_')
		}
	}

	return string(result)
}

// GinHandler returns a Gin handler function for serving Prometheus metrics via HTTP.
// The handler exposes all metrics from the collector's registry in Prometheus text format.
// It supports OpenMetrics format for enhanced metric capabilities.
// Example usage:
//
//	router := gin.Default()
//	collector := NewPrometheusCollector()
//	router.GET("/metrics", collector.GinHandler())
//
// This function is useful for adding a metrics endpoint to your Gin-based HTTP server,
// allowing Prometheus to scrape metrics from your application.
func (p *PrometheusCollector) GinHandler() gin.HandlerFunc {
	handler := promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})

	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}

// PrometheusHandler returns a Gin handler function for serving metrics from a specific Prometheus registry.
// This standalone function allows you to expose metrics from any Prometheus registry via HTTP.
// Example usage:
//
//	router := gin.Default()
//	reg := prometheus.NewRegistry()
//	router.GET("/metrics", PrometheusHandler(reg))
//
// This function is useful when you have a custom registry that you want to expose
// independently of a PrometheusCollector instance.
func PrometheusHandler(registry *prometheus.Registry) gin.HandlerFunc {
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})

	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}

// AllPrometheusHandler returns a Gin handler function for serving all metrics from the default Prometheus registry.
// This includes metrics from the default registry which may contain Go runtime metrics and other default collectors.
// Example usage:
//
//	router := gin.Default()
//	router.GET("/metrics", AllPrometheusHandler())
//
// This function is useful when you want to expose all Prometheus metrics in your application,
// including both custom metrics and default Go metrics, without creating a separate collector.
func AllPrometheusHandler() gin.HandlerFunc {
	handler := promhttp.Handler()

	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}
