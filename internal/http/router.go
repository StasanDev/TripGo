package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	api "github.com/StasanDev/avitoLabs/internal/generated"
	"github.com/StasanDev/avitoLabs/internal/http/middleware"
)

func NewRouter(
	handler api.ServerInterface,
	logger *slog.Logger,
	tracerProvider trace.TracerProvider,
	meterProvider metric.MeterProvider,
) (http.Handler, error) {
	validator, err := middleware.RequestValidator(logger)
	if err != nil {
		return nil, err
	}
	r := chi.NewRouter()
	metrics, err := middleware.Metrics(meterProvider, r)
	if err != nil {
		return nil, err
	}

	r.Use(
		middleware.Recover(logger),
		middleware.Tracing(tracerProvider),
		middleware.Logging(logger),
		metrics,
	)

	return api.HandlerWithOptions(handler, api.ChiServerOptions{
		BaseRouter:       r,
		Middlewares:      []api.MiddlewareFunc{validator},
		ErrorHandlerFunc: middleware.ValidationErrorHandler(logger),
	}), nil
}
