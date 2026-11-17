package database

import (
	"context"
	"strings"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.opentelemetry.io/otel/trace"
)

type dbOperation struct {
	startedAt  time.Time
	name       string
	collection string
}

type dbOperationContextKey struct{}

type dbTracer struct {
	*otelpgx.Tracer
	operationDuration metric.Float64Histogram
}

func newDBTracer(
	tracerProvider trace.TracerProvider,
	meter metric.Meter,
) (*dbTracer, error) {
	duration, err := meter.Float64Histogram(
		"db.client.operation.duration",
		metric.WithDescription("Duration of database client operations."),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(
			0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10,
		),
	)
	if err != nil {
		return nil, err
	}

	return &dbTracer{
		Tracer: otelpgx.NewTracer(
			otelpgx.WithTracerProvider(tracerProvider),
			otelpgx.WithMeterProvider(metricnoop.NewMeterProvider()),
		),
		operationDuration: duration,
	}, nil
}

func (t *dbTracer) TraceQueryStart(
	ctx context.Context,
	conn *pgx.Conn,
	data pgx.TraceQueryStartData,
) context.Context {
	ctx = t.Tracer.TraceQueryStart(ctx, conn, data)

	operation := dbOperation{
		startedAt:  time.Now(),
		name:       otelpgx.SQLOperationName(ctx, data.SQL),
		collection: collectionFromSQL(data.SQL),
	}
	if operation.collection != "" {
		trace.SpanFromContext(ctx).SetAttributes(
			semconv.DBCollectionName(operation.collection),
		)
	}

	return context.WithValue(ctx, dbOperationContextKey{}, operation)
}

func (t *dbTracer) TraceQueryEnd(
	ctx context.Context,
	conn *pgx.Conn,
	data pgx.TraceQueryEndData,
) {
	operation, ok := ctx.Value(dbOperationContextKey{}).(dbOperation)
	if ok {
		attrs := make([]attribute.KeyValue, 0, 3)
		if operation.name != "" {
			attrs = append(attrs, semconv.DBOperationName(operation.name))
		}
		if operation.collection != "" {
			attrs = append(attrs, semconv.DBCollectionName(operation.collection))
		}
		if data.Err != nil {
			errorType := semconv.ErrorType(data.Err)
			attrs = append(attrs, errorType)
			trace.SpanFromContext(ctx).SetAttributes(errorType)
		}

		t.operationDuration.Record(
			ctx,
			time.Since(operation.startedAt).Seconds(),
			metric.WithAttributes(attrs...),
		)
	}

	t.Tracer.TraceQueryEnd(ctx, conn, data)
}

func collectionFromSQL(query string) string {
	fields := strings.Fields(query)
	if len(fields) < 2 {
		return ""
	}

	operation := strings.ToUpper(strings.Trim(fields[0], "(\"`"))
	var collection string

	switch operation {
	case "INSERT":
		collection = fieldAfter(fields, "INTO")
	case "SELECT", "DELETE":
		collection = fieldAfter(fields, "FROM")
	case "UPDATE":
		collection = fields[1]
	case "COPY":
		collection = fields[1]
	}

	return normalizeCollection(collection)
}

func fieldAfter(fields []string, keyword string) string {
	for i := 1; i < len(fields)-1; i++ {
		if strings.EqualFold(strings.Trim(fields[i], "(\"`"), keyword) {
			return fields[i+1]
		}
	}

	return ""
}

func normalizeCollection(collection string) string {
	collection = strings.TrimSpace(collection)
	if collection == "" || strings.HasPrefix(collection, "(") {
		return ""
	}

	parts := strings.Split(collection, ".")
	collection = parts[len(parts)-1]

	return strings.Trim(collection, "\"`[](),;")
}
