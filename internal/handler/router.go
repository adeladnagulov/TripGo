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
	r.Get("/ready", h.Ready)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/trips", h.CreateTrip)
	})

	return r
}
