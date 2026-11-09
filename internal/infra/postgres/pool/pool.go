package database

import (
	"context"
	"fmt"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/StasanDev/avitoLabs/internal/config"
)

const dbInstrumentationScope = "trip-service/db"

func NewPostgresPool(
	ctx context.Context,
	cfg *config.DBConfig,
	tracerProvider trace.TracerProvider,
	meterProvider metric.MeterProvider,
) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.Url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	dbTracer, err := newDBTracer(
		tracerProvider,
		meterProvider.Meter(dbInstrumentationScope),
	)
	if err != nil {
		return nil, fmt.Errorf("create database telemetry: %w", err)
	}
	poolCfg.ConnConfig.Tracer = dbTracer

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	poolCfg.ConnConfig.ConnectTimeout = cfg.ConnectTimeout

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	if err := otelpgx.RecordStats(
		pool,
		otelpgx.WithStatsMeterProvider(meterProvider),
	); err != nil {
		pool.Close()
		return nil, fmt.Errorf("record postgres pool stats: %w", err)
	}

	return pool, nil
}
