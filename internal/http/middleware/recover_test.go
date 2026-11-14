package middleware_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric/noop"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	api "github.com/StasanDev/avitoLabs/internal/generated"
	"github.com/StasanDev/avitoLabs/internal/http/middleware"
	"github.com/StasanDev/avitoLabs/internal/infra/telemetry"
)

func TestRecover(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	handler := middleware.Recover(logger)(http.HandlerFunc(
		func(http.ResponseWriter, *http.Request) {
			panic("sensitive panic text")
		},
	))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/panic", nil)

	require.NotPanics(t, func() {
		handler.ServeHTTP(recorder, request)
	})

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Equal(t, "application/problem+json", recorder.Header().Get("Content-Type"))

	var problem api.Problem
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&problem))
	assert.Equal(t, "internal_error", problem.Code)
	assert.NotContains(t, recorder.Body.String(), "sensitive panic text")

	var logRecord map[string]any
	require.NoError(t, json.NewDecoder(&logs).Decode(&logRecord))
	assert.Equal(t, "ERROR", logRecord["level"])
	assert.Equal(t, "panic recovered", logRecord["msg"])
	assert.Equal(t, "sensitive panic text", logRecord["panic"])
	assert.Contains(t, logRecord["stack"], "TestRecover")
}

func TestRecoverPanicLogsShareHTTPSpan(t *testing.T) {
	var logs bytes.Buffer
	logger, err := telemetry.NewLogger(&logs, "INFO", "trip-service")
	require.NoError(t, err)

	tracerProvider := sdktrace.NewTracerProvider()
	t.Cleanup(func() {
		require.NoError(t, tracerProvider.Shutdown(context.Background()))
	})

	router := chi.NewRouter()
	metrics, err := middleware.Metrics(noop.NewMeterProvider(), router)
	require.NoError(t, err)
	router.Use(
		middleware.Recover(logger),
		middleware.Tracing(tracerProvider),
		middleware.Logging(logger),
		metrics,
	)
	var spanContext trace.SpanContext
	router.Get("/panic", func(_ http.ResponseWriter, r *http.Request) {
		spanContext = trace.SpanContextFromContext(r.Context())
		panic("sensitive panic text")
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/panic", nil)
	router.ServeHTTP(recorder, request)
	require.True(t, spanContext.IsValid())

	decoder := json.NewDecoder(&logs)
	var requestLog, panicLog map[string]any
	require.NoError(t, decoder.Decode(&requestLog))
	require.NoError(t, decoder.Decode(&panicLog))

	assert.Equal(t, "request handled", requestLog["msg"])
	assert.Equal(t, "panic recovered", panicLog["msg"])
	assert.NotEmpty(t, panicLog["stack"])

	for _, record := range []map[string]any{requestLog, panicLog} {
		assert.Equal(t, spanContext.TraceID().String(), record["trace_id"])
		assert.Equal(t, spanContext.SpanID().String(), record["span_id"])
	}
}
