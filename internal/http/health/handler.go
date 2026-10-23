package health

import (
	"context"
	"net/http"

	"github.com/adeladnagulov/TripGo/api"
	"github.com/adeladnagulov/TripGo/internal/http/domainerror"
	"github.com/adeladnagulov/TripGo/internal/http/response"
)

type DbPinger interface {
	Ping(ctx context.Context) error
}

type Handler struct {
	db DbPinger
}

func NewHandler(db DbPinger) *Handler {
	return &Handler{
		db: db,
	}
}

// Health Liveness
// (GET /health)
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	resp := api.HealthResponse{
		Status: api.Ok,
	}
	response.Json(w, http.StatusOK, resp)
}

// Ready Readiness
// (GET /ready)
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	if err := h.db.Ping(r.Context()); err != nil {
		status, problem := domainerror.Problem(err, r)
		response.Problem(w, status, problem)
		return
	}
	resp := api.HealthResponse{
		Status: api.Ok,
	}
	response.Json(w, http.StatusOK, resp)
}
