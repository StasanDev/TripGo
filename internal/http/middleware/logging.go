package middleware

import (
	"log/slog"
	"net/http"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()
			responseWriter := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

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

				attributes := []slog.Attr{
					slog.String("method", r.Method),
					slog.String("route", routePattern(r)),
					slog.Int("status", status),
					slog.Int64("duration_ms", time.Since(startedAt).Milliseconds()),
				}

				if recovered != nil || status >= http.StatusInternalServerError {
					logger.LogAttrs(r.Context(), slog.LevelError, "request handled", attributes...)
				} else {
					logger.LogAttrs(r.Context(), slog.LevelInfo, "request handled", attributes...)
				}

				if recovered != nil {
					panic(recovered)
				}
			}()

			next.ServeHTTP(responseWriter, r)
		})
	}
}
