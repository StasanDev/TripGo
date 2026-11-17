package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/StasanDev/avitoLabs/internal/http/middleware"
)

func TestTracing(t *testing.T) {
	spanRecorder := tracetest.NewSpanRecorder()
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(spanRecorder),
	)
	t.Cleanup(func() {
		require.NoError(t, tracerProvider.Shutdown(context.Background()))
	})

	router := chi.NewRouter()
	router.Use(middleware.Tracing(tracerProvider))
	router.Get("/trips/{tripId}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/trips/trip-123", nil)
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Len(t, spanRecorder.Ended(), 1)

	span := spanRecorder.Ended()[0]
	assert.Equal(t, "GET /trips/{tripId}", span.Name())

	attributes := make(map[string]any, len(span.Attributes()))
	for _, attribute := range span.Attributes() {
		attributes[string(attribute.Key)] = attribute.Value.AsInterface()
	}
	assert.Equal(t, "/trips/{tripId}", attributes["http.route"])
}
