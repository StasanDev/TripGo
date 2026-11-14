package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"

	"go.opentelemetry.io/otel/trace"

	"github.com/StasanDev/avitoLabs/internal/http/response"
)

var errRecoveredPanic = errors.New("recovered panic")

type requestStateKey struct{}

type requestState struct {
	spanContext trace.SpanContext
}

func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			state := &requestState{}
			ctx := context.WithValue(r.Context(), requestStateKey{}, state)
			r = r.WithContext(ctx)
			defer func() {
				recovered := recover()
				if recovered == nil {
					return
				}

				if state.spanContext.IsValid() {
					ctx = trace.ContextWithSpanContext(ctx, state.spanContext)
				}

				logger.ErrorContext(
					ctx,
					"panic recovered",
					slog.Any("panic", recovered),
					slog.String("stack", string(debug.Stack())),
				)

				_, err := response.WriteError(w, r, errRecoveredPanic)
				if err != nil {
					logger.ErrorContext(
						ctx,
						"failed to write panic response",
						slog.Any("error", err),
					)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
