package get

import (
	"net/http"

	"github.com/adeladnagulov/TripGo/api"
	"github.com/adeladnagulov/TripGo/internal/http/domainerror"
	"github.com/adeladnagulov/TripGo/internal/http/helper/response"
)

type Handler struct {
	uc tripUseCase
}

func NewHandler(uc tripUseCase) *Handler {
	return &Handler{uc: uc}
}

// GetTrip Получить поездку
// (GET /api/v1/trips/{tripId})
func (h *Handler) GetTrip(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	t, err := h.uc.GetTrip(r.Context(), tripId)
	if err != nil {
		status, problem := domainerror.Problem(err, r)
		response.Problem(w, status, problem)
		return
	}
	resp := toApi(*t)
	response.Json(w, http.StatusOK, resp)
}
