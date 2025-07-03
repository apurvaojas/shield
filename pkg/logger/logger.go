package logger

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

type Logger struct {
	*logrus.Logger
}

func New() *Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)

	return &Logger{Logger: log}
}

func (l *Logger) WithContext(ctx context.Context) *logrus.Entry {
	entry := l.WithFields(logrus.Fields{})

	// Add trace context if available
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		spanCtx := span.SpanContext()
		entry = entry.WithFields(logrus.Fields{
			"trace_id": spanCtx.TraceID().String(),
			"span_id":  spanCtx.SpanID().String(),
		})
	}

	return entry
}

func (l *Logger) WithCorrelationID(correlationID string) *logrus.Entry {
	return l.WithField("correlation_id", correlationID)
}