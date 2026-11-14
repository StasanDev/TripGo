package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/StasanDev/avitoLabs/internal/http/response"
)

var errRecoveredPanic = errors.New("recovered panic")

func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				recovered := recover()
				if recovered == nil {
					return
				}

				logger.ErrorContext(
					r.Context(),
					"panic recovered",
					slog.Any("panic", recovered),
					slog.String("stack", string(debug.Stack())),
				)

				_, err := response.WriteError(w, r, errRecoveredPanic)
				if err != nil {
					logger.ErrorContext(
						r.Context(),
						"failed to write panic response",
						slog.Any("error", err),
					)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
