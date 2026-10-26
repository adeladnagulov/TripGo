package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(lrw, r)

		duration := time.Since(start)
		level := slog.LevelInfo
		if lrw.status >= 500 {
			level = slog.LevelError
		}

		slog.LogAttrs(r.Context(), level, "request handled",
			slog.String("service", "trip-service"),
			slog.String("metod", r.Method),
			slog.String("route", r.URL.Path),
			slog.Int("status", lrw.status),
			slog.Int64("duration_ms", duration.Milliseconds()),
		)
	})
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
