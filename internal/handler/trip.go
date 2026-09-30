package handler

import (
	"net/http"

	"github.com/adeladnagulov/TripGo/internal/usecase"
)

type Handler struct {
	tripServise *usecase.TripServise
}

func newHandler(tripServise *usecase.TripServise) *Handler {
	return &Handler{
		tripServise: tripServise,
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	if err := h.tripServise.PingTrip(r.Context()); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status": "error", "reason": "database_unreachable"}`))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ready"}`))
}
