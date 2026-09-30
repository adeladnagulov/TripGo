package handler

import (
	"net/http"

	"github.com/adeladnagulov/TripGo/internal/usecase"
	"github.com/go-chi/chi/v5"
)

func NewRouter(tripServise *usecase.TripServise) http.Handler {
	r := chi.NewRouter()

	h := newHandler(tripServise)

	r.Get("/health", h.Health)

	return r
}
