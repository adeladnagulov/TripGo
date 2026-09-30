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
