package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func Metrics(meter metric.Meter) func(http.Handler) http.Handler {
	activeReq, _ := meter.Int64UpDownCounter(
		"http.server.active_requests",
		metric.WithUnit("{request}"),
		metric.WithDescription("how many requests are being processed right now"),
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			routePattern := chi.RouteContext(r.Context()).RoutePattern()
			activeReq.Add(r.Context(), 1, metric.WithAttributes(
				attribute.String("http.request.method", r.Method),
				attribute.String("http.route", routePattern),
			))

			defer func() {
				if p := recover(); p != nil {
					panic(p)
				}
				activeReq.Add(r.Context(), -1, metric.WithAttributes(
					attribute.String("http.request.method", r.Method),
					attribute.String("http.route", routePattern),
				))
			}()
			lrw := &statusResponseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(lrw, r)
		})
	}
}
