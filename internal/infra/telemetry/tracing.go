package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func newTracerProvider(
	ctx context.Context,
	res *resource.Resource,
) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithTimeout(2*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("create OTLP trace exporter: %w", err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(
			exporter,
			sdktrace.WithBatchTimeout(time.Second),
		),
	)

	return provider, nil
}
