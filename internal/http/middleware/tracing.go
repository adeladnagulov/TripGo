package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func Tracing() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				routePattern := chi.RouteContext(r.Context()).RoutePattern()
				span := trace.SpanFromContext(r.Context())
				span.SetAttributes(attribute.String("http.route", routePattern))
				next.ServeHTTP(w, r)
			}),
			"server",
			otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
				routePattern := chi.RouteContext(r.Context()).RoutePattern()
				return r.Method + " " + routePattern
			}),
		)
	}
}
