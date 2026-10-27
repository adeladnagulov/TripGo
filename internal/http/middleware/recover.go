package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/adeladnagulov/TripGo/internal/http/domainerror"
	"github.com/adeladnagulov/TripGo/internal/http/response"
)

func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.ErrorContext(r.Context(), "panic recovered",
						"error", err,
						"stack", string(debug.Stack()),
					)

					status, problem := domainerror.Problem(errors.New("internal server error"), r)
					response.Problem(w, status, problem)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
