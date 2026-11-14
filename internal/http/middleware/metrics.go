package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

const meterName = "github.com/StasanDev/avitoLabs/internal/http/middleware"

type httpMetrics struct {
	routes          chi.Routes
	requestDuration metric.Float64Histogram
	activeRequests  metric.Int64UpDownCounter
}

func Metrics(
	meterProvider metric.MeterProvider,
	routes chi.Routes,
) (func(http.Handler) http.Handler, error) {
	meter := meterProvider.Meter(meterName)

	requestDuration, err := meter.Float64Histogram(
		"http.server.request.duration",
		metric.WithDescription("Duration of HTTP server requests."),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("create HTTP request duration metric: %w", err)
	}

	activeRequests, err := meter.Int64UpDownCounter(
		"http.server.active_requests",
		metric.WithDescription("Number of active HTTP server requests."),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return nil, fmt.Errorf("create active HTTP requests metric: %w", err)
	}

	metrics := &httpMetrics{
		routes:          routes,
		requestDuration: requestDuration,
		activeRequests:  activeRequests,
	}

	return metrics.middleware, nil
}

func (m *httpMetrics) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		responseWriter := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
		route := matchRoute(m.routes, r)
		requestAttributes := []attribute.KeyValue{
			semconv.HTTPRequestMethodKey.String(r.Method),
			semconv.HTTPRoute(route),
		}
		requestOptions := metric.WithAttributes(requestAttributes...)

		m.activeRequests.Add(r.Context(), 1, requestOptions)

		defer func() {
			recovered := recover()
			status := responseWriter.Status()
			if status == 0 {
				if recovered != nil {
					status = http.StatusInternalServerError
				} else {
					status = http.StatusOK
				}
			}

			m.activeRequests.Add(r.Context(), -1, requestOptions)
			m.requestDuration.Record(
				r.Context(),
				time.Since(startedAt).Seconds(),
				metric.WithAttributes(
					append(requestAttributes, semconv.HTTPResponseStatusCode(status))...,
				),
			)

			if recovered != nil {
				panic(recovered)
			}
		}()

		next.ServeHTTP(responseWriter, r)
	})
}

func matchRoute(routes chi.Routes, r *http.Request) string {
	routeContext := chi.NewRouteContext()
	routePath := r.URL.RawPath
	if routePath == "" {
		routePath = r.URL.Path
	}

	if !routes.Match(routeContext, r.Method, routePath) {
		return ""
	}

	return routeContext.RoutePattern()
}
