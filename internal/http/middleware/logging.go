package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lrw := &loggingResponseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(lrw, r)

			duration := time.Since(start)
			level := slog.LevelInfo
			if lrw.status >= 500 {
				level = slog.LevelError
			}

			routePattern := ""
			if routeCtx := chi.RouteContext(r.Context()); routeCtx != nil {
				routePattern = routeCtx.RoutePattern()
			}

			logger.LogAttrs(r.Context(), level, "request handled",
				slog.String("service", "trip-service"),
				slog.String("method", r.Method),
				slog.String("route", routePattern),
				slog.Int("status", lrw.status),
				slog.Int64("duration_ms", duration.Milliseconds()),
			)
		})
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	if !lrw.wroteHeader {
		lrw.status = code
		lrw.wroteHeader = true
		lrw.ResponseWriter.WriteHeader(code)
	}
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	if !lrw.wroteHeader {
		lrw.WriteHeader(http.StatusOK)
	}
	return lrw.ResponseWriter.Write(b)
}
