// Package instrumentation provides OpenTelemetry (OTEL) middleware for Gin applications.
// This middleware provides distributed tracing, metrics, and spans for HTTP requests,
// enabling comprehensive observability for microservices architectures.
package instrumentation

import (
	"strings"

	appconfig "shield/cmd/app/config"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// OTELMiddlewareConfig holds configuration for OTEL middleware
type OTELMiddlewareConfig struct {
	ServiceName         string
	EnableTracing       bool
	EnableMetrics       bool
	WithSpanID          bool
	WithTraceID         bool
	WithUserAgent       bool
	WithRequestBody     bool
	WithResponseBody    bool
	WithRequestHeader   bool
	WithResponseHeader  bool
	CustomSpanFormatter otelgin.SpanNameFormatter
	TracerProvider      trace.TracerProvider
	MeterProvider       metric.MeterProvider
	Propagators         propagation.TextMapPropagator
	SpanStartOptions    []trace.SpanStartOption
	FilterPaths         []string
	FilterMethods       []string
}

// InitOTELMiddleware initializes and returns the OpenTelemetry middleware for Gin
// using centralized configuration. If config is nil, default configuration is used.
func InitOTELMiddleware(config *OTELMiddlewareConfig) gin.HandlerFunc {
	viperConfig := appconfig.GetInstrumentationConfig().OpenTelemetry

	if config == nil {
		// Convert Viper config to middleware config
		config = &OTELMiddlewareConfig{
			ServiceName:         viperConfig.ServiceName,
			EnableTracing:       viperConfig.EnableTracing,
			EnableMetrics:       viperConfig.EnableMetrics,
			WithSpanID:          viperConfig.WithSpanID,
			WithTraceID:         viperConfig.WithTraceID,
			WithUserAgent:       viperConfig.WithUserAgent,
			WithRequestBody:     viperConfig.WithRequestBody,
			WithResponseBody:    viperConfig.WithResponseBody,
			WithRequestHeader:   viperConfig.WithRequestHeader,
			WithResponseHeader:  viperConfig.WithResponseHeader,
			FilterPaths:         viperConfig.FilterPaths,
			FilterMethods:       viperConfig.FilterMethods,
			TracerProvider:      nil, // Will use global provider if not set
			MeterProvider:       nil, // Will use global provider if not set
			Propagators:         nil, // Will use global propagators if not set
			SpanStartOptions:    []trace.SpanStartOption{},
			CustomSpanFormatter: nil, // Will use default formatter if not set
		}
	}

	// Skip if tracing is disabled
	if !config.EnableTracing {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// Build otelgin options
	var opts []otelgin.Option

	// Set tracer provider if provided
	if config.TracerProvider != nil {
		opts = append(opts, otelgin.WithTracerProvider(config.TracerProvider))
	}

	// Set meter provider if provided
	if config.MeterProvider != nil {
		opts = append(opts, otelgin.WithMeterProvider(config.MeterProvider))
	}

	// Set propagators if provided
	if config.Propagators != nil {
		opts = append(opts, otelgin.WithPropagators(config.Propagators))
	}

	// Add span start options if provided
	if len(config.SpanStartOptions) > 0 {
		opts = append(opts, otelgin.WithSpanStartOptions(config.SpanStartOptions...))
	}

	// Set custom span name formatter if provided
	if config.CustomSpanFormatter != nil {
		opts = append(opts, otelgin.WithSpanNameFormatter(config.CustomSpanFormatter))
	}

	// Add filters for paths
	if len(config.FilterPaths) > 0 {
		opts = append(opts, otelgin.WithGinFilter(func(c *gin.Context) bool {
			path := c.Request.URL.Path
			for _, filterPath := range config.FilterPaths {
				if strings.Contains(path, filterPath) {
					return false // Skip this path
				}
			}
			return true // Allow this path
		}))
	}

	// Add filters for methods
	if len(config.FilterMethods) > 0 {
		opts = append(opts, otelgin.WithGinFilter(func(c *gin.Context) bool {
			method := c.Request.Method
			for _, filterMethod := range config.FilterMethods {
				if method == filterMethod {
					return false // Skip this method
				}
			}
			return true // Allow this method
		}))
	}
	// Add metric attribute function for enhanced metrics with sensitive data masking
	if config.EnableMetrics {
		opts = append(opts, otelgin.WithGinMetricAttributeFn(func(c *gin.Context) []attribute.KeyValue {
			masker := GetDefaultMasker()

			attrs := []attribute.KeyValue{
				attribute.String("http.route", c.FullPath()),
			}

			// Mask User-Agent if it contains PII
			if userAgent := c.GetHeader("User-Agent"); userAgent != "" {
				maskedUserAgent := masker.MaskPII(userAgent)
				attrs = append(attrs, attribute.String("http.user_agent", maskedUserAgent))
			}

			// Add request ID if available (safe to include)
			if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
				attrs = append(attrs, attribute.String("http.request_id", requestID))
			}

			// Mask query parameters if present
			if c.Request.URL.RawQuery != "" {
				maskedQuery := masker.MaskQueryParams(c.Request.URL.RawQuery)
				attrs = append(attrs, attribute.String("http.query", maskedQuery))
			}

			// Add custom attributes from context (already masked when added)
			if val, exists := c.Get("otel.custom_attributes"); exists {
				if customAttrs, ok := val.([]attribute.KeyValue); ok {
					attrs = append(attrs, customAttrs...)
				}
			}

			return attrs
		}))
	}

	return otelgin.Middleware(config.ServiceName, opts...)
}

// InitOTELMiddlewareDefault initializes the OpenTelemetry middleware with Viper configuration
func InitOTELMiddlewareDefault() gin.HandlerFunc {
	return InitOTELMiddleware(nil)
}

// AddOTELAttributes adds custom attributes to the current span and metrics with automatic PII masking
// This function can be used within request handlers to add context-specific telemetry
func AddOTELAttributes(c *gin.Context, attrs ...attribute.KeyValue) {
	masker := GetDefaultMasker()
	maskedAttrs := make([]attribute.KeyValue, len(attrs))

	// Mask PII in attribute values
	for i, attr := range attrs {
		switch attr.Value.Type() {
		case attribute.STRING:
			maskedValue := masker.MaskPII(attr.Value.AsString())
			maskedAttrs[i] = attribute.String(string(attr.Key), maskedValue)
		default:
			maskedAttrs[i] = attr
		}
	}

	// Add to span if span is available
	span := trace.SpanFromContext(c.Request.Context())
	if span != nil && span.IsRecording() {
		span.SetAttributes(maskedAttrs...)
	}

	// Store for metrics collection
	if existing, exists := c.Get("otel.custom_attributes"); exists {
		if existingAttrs, ok := existing.([]attribute.KeyValue); ok {
			maskedAttrs = append(existingAttrs, maskedAttrs...)
		}
	}
	c.Set("otel.custom_attributes", maskedAttrs)
}

// AddOTELAttributesUnsafe adds custom attributes without PII masking
// Use this only when you're certain the attributes don't contain sensitive data
func AddOTELAttributesUnsafe(c *gin.Context, attrs ...attribute.KeyValue) {
	// Add to span if span is available
	span := trace.SpanFromContext(c.Request.Context())
	if span != nil && span.IsRecording() {
		span.SetAttributes(attrs...)
	}

	// Store for metrics collection
	if existing, exists := c.Get("otel.custom_attributes"); exists {
		if existingAttrs, ok := existing.([]attribute.KeyValue); ok {
			attrs = append(existingAttrs, attrs...)
		}
	}
	c.Set("otel.custom_attributes", attrs)
}

// GetTraceID returns the current trace ID from the request context
func GetTraceID(c *gin.Context) string {
	span := trace.SpanFromContext(c.Request.Context())
	if span != nil {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// GetSpanID returns the current span ID from the request context
func GetSpanID(c *gin.Context) string {
	span := trace.SpanFromContext(c.Request.Context())
	if span != nil {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// StartChildSpan starts a new child span for the current request
// This is useful for creating custom spans for specific operations within a request
func StartChildSpan(c *gin.Context, operationName string, opts ...trace.SpanStartOption) (trace.Span, func()) {
	tracer := otel.Tracer("gin-custom-span")
	ctx, span := tracer.Start(c.Request.Context(), operationName, opts...)

	// Update the context in gin.Context
	c.Request = c.Request.WithContext(ctx)

	return span, func() {
		span.End()
	}
}

// RecordError records an error in the current span
func RecordError(c *gin.Context, err error, opts ...trace.EventOption) {
	span := trace.SpanFromContext(c.Request.Context())
	if span != nil && span.IsRecording() {
		span.RecordError(err, opts...)
	}
}

// SetSpanStatus sets the status of the current span
func SetSpanStatus(c *gin.Context, code codes.Code, description string) {
	span := trace.SpanFromContext(c.Request.Context())
	if span != nil && span.IsRecording() {
		span.SetStatus(code, description)
	}
}
