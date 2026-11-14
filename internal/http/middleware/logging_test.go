package middleware_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/StasanDev/avitoLabs/internal/http/middleware"
	"github.com/StasanDev/avitoLabs/internal/infra/telemetry"
)

func TestLogging(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		wantLevel string
	}{
		{
			name:      "success",
			status:    http.StatusCreated,
			wantLevel: "INFO",
		},
		{
			name:      "client error",
			status:    http.StatusBadRequest,
			wantLevel: "INFO",
		},
		{
			name:      "server error",
			status:    http.StatusInternalServerError,
			wantLevel: "ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			logger, err := telemetry.NewLogger(&logs, "debug", "trip-service")
			require.NoError(t, err)

			tracerProvider := sdktrace.NewTracerProvider()
			t.Cleanup(func() {
				require.NoError(t, tracerProvider.Shutdown(context.Background()))
			})

			router := chi.NewRouter()
			router.Use(
				middleware.Tracing(tracerProvider),
				middleware.Logging(logger),
			)
			router.Post("/trips/{tripId}/finish", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
			})

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/trips/trip-123/finish", nil)
			router.ServeHTTP(recorder, request)

			require.Equal(t, tt.status, recorder.Code)

			var record map[string]any
			require.NoError(t, json.NewDecoder(&logs).Decode(&record))
			assert.Equal(t, tt.wantLevel, record["level"])
			assert.Equal(t, "request handled", record["msg"])
			assert.Equal(t, "trip-service", record["service"])
			assert.Equal(t, http.MethodPost, record["method"])
			assert.Equal(t, "/trips/{tripId}/finish", record["route"])
			assert.Equal(t, float64(tt.status), record["status"])
			assert.IsType(t, float64(0), record["duration_ms"])
			assert.NotEmpty(t, record["trace_id"])
			assert.NotEmpty(t, record["span_id"])
			assert.NotContains(t, logs.String(), "trip-123")
		})
	}
}

func TestLoggingRethrowsPanicForRecover(t *testing.T) {
	var logs bytes.Buffer
	logger, err := telemetry.NewLogger(&logs, "debug", "trip-service")
	require.NoError(t, err)

	handler := middleware.Logging(logger)(http.HandlerFunc(
		func(http.ResponseWriter, *http.Request) {
			panic("panic value")
		},
	))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/panic", nil)

	assert.PanicsWithValue(t, "panic value", func() {
		handler.ServeHTTP(recorder, request)
	})

	var record map[string]any
	require.NoError(t, json.NewDecoder(&logs).Decode(&record))
	assert.Equal(t, "ERROR", record["level"])
	assert.Equal(t, float64(http.StatusInternalServerError), record["status"])
}
