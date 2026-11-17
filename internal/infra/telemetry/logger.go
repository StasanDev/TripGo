package telemetry

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

func NewLogger(out io.Writer, levelText, service string) (*slog.Logger, error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(levelText)); err != nil {
		return nil, fmt.Errorf("parse log level: %w", err)
	}

	handler := slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level: level,
	})

	return slog.New(traceHandler{
		Handler: handler,
	}).With(
		slog.String("service", service),
	), nil
}

type traceHandler struct {
	slog.Handler
}

func (h traceHandler) Handle(ctx context.Context, record slog.Record) error {
	var traceID, spanID string

	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.IsValid() {
		traceID = spanContext.TraceID().String()
		spanID = spanContext.SpanID().String()
	}

	record.AddAttrs(
		slog.String("trace_id", traceID),
		slog.String("span_id", spanID),
	)

	return h.Handler.Handle(ctx, record)
}

func (h traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return traceHandler{
		Handler: h.Handler.WithAttrs(attrs),
	}
}

func (h traceHandler) WithGroup(name string) slog.Handler {
	return traceHandler{
		Handler: h.Handler.WithGroup(name),
	}
}
