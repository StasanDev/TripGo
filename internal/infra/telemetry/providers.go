package telemetry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const instrumentationScope = "github.com/StasanDev/avitoLabs/internal/infra/telemetry"

type Providers struct {
	Tracer *sdktrace.TracerProvider
	Meter  *sdkmetric.MeterProvider
	logger *slog.Logger

	created   metric.Int64Counter
	completed metric.Int64Counter
}

func (p *Providers) Create(ctx context.Context, tripID uuid.UUID) {
	p.created.Add(ctx, 1)
	p.logger.InfoContext(
		ctx,
		"trip created",
		slog.String("trip_id", tripID.String()),
	)
}

func (p *Providers) Complete(ctx context.Context, tripID uuid.UUID) {
	p.completed.Add(ctx, 1)
	p.logger.InfoContext(
		ctx,
		"trip completed",
		slog.String("trip_id", tripID.String()),
	)
}

func NewProviders(
	ctx context.Context,
	serviceName string,
	serviceVersion string,
	timeout time.Duration,
	logger *slog.Logger,
) (*Providers, error) {
	res, err := newResource(ctx, serviceName, serviceVersion)
	if err != nil {
		return nil, err
	}

	tracerProvider, err := newTracerProvider(ctx, res)
	if err != nil {
		return nil, err
	}

	meterProvider, err := newMeterProvider(ctx, res)
	if err != nil {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		return nil, errors.Join(
			err,
			tracerProvider.Shutdown(cleanupCtx),
		)
	}

	providers := &Providers{
		Tracer: tracerProvider,
		Meter:  meterProvider,
		logger: logger,
	}
	providers.created, err = providers.Meter.Meter(instrumentationScope).Int64Counter(
		"tripgo.trips.created",
		metric.WithDescription("Number of created trips."),
		metric.WithUnit("{trip}"),
	)
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("create trips created counter: %w", err),
			shutdownProviders(providers, timeout),
		)
	}

	providers.completed, err = providers.Meter.Meter(instrumentationScope).Int64Counter(
		"tripgo.trips.completed",
		metric.WithDescription("Number of completed trips."),
		metric.WithUnit("{trip}"),
	)
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("create trips completed counter: %w", err),
			shutdownProviders(providers, timeout),
		)
	}

	otel.SetTracerProvider(providers.Tracer)
	otel.SetMeterProvider(providers.Meter)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return providers, nil
}

func shutdownProviders(providers *Providers, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return providers.Shutdown(ctx)
}

func (p *Providers) Shutdown(ctx context.Context) error {
	var result error

	if err := p.Meter.Shutdown(ctx); err != nil {
		result = errors.Join(result, fmt.Errorf("shutdown metrics: %w", err))
	}
	if err := p.Tracer.Shutdown(ctx); err != nil {
		result = errors.Join(result, fmt.Errorf("shutdown tracing: %w", err))
	}

	return result
}
