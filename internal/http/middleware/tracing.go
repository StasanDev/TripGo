package middleware

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

func Tracing(tracerProvider trace.TracerProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		tracedHandler := otelhttp.NewMiddleware(
			"http.server",
			otelhttp.WithTracerProvider(tracerProvider),
			// HTTP metrics are collected by the separate Metrics middleware.
			otelhttp.WithMeterProvider(noop.NewMeterProvider()),
		)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			responseWriter := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			defer func() {
				recovered := recover()
				if recovered != nil {
					markHTTPSpanPanic(r)
					panic(recovered)
				}

				enrichHTTPSpan(r, responseWriter.Status())
			}()

			next.ServeHTTP(responseWriter, r)
		}))

		return tracedHandler
	}
}

func enrichHTTPSpan(r *http.Request, status int) {
	span := trace.SpanFromContext(r.Context())

	route := routePattern(r)
	if route != "" {
		span.SetName(r.Method + " " + route)
		span.SetAttributes(semconv.HTTPRoute(route))
	}

	if status >= http.StatusInternalServerError {
		span.SetAttributes(
			semconv.ErrorTypeKey.String(strconv.Itoa(status)),
		)
	}
}

func markHTTPSpanPanic(r *http.Request) {
	span := trace.SpanFromContext(r.Context())
	span.RecordError(errRecoveredPanic, trace.WithStackTrace(true))
	span.SetStatus(codes.Error, errRecoveredPanic.Error())
	span.SetAttributes(
		semconv.HTTPResponseStatusCode(http.StatusInternalServerError),
		semconv.ErrorTypeKey.String("panic"),
	)

	route := routePattern(r)
	if route != "" {
		span.SetName(r.Method + " " + route)
		span.SetAttributes(semconv.HTTPRoute(route))
	}
}

func routePattern(r *http.Request) string {
	routeContext := chi.RouteContext(r.Context())
	if routeContext == nil {
		return ""
	}

	return routeContext.RoutePattern()
}
