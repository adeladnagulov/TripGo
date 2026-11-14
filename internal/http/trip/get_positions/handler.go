package getpositions

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

// ListTripPositions Получить маршрут поездки
// (GET /api/v1/trips/{tripId}/positions)
func (h *Handler) ListTripPositions(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	positions, err := h.uc.GetPositions(r.Context(), tripId)
	if err != nil {
		status, problem := domainerror.Problem(err, r)
		response.Problem(w, status, problem)
		return
	}
	resp := toApi(*positions)
	response.Json(w, http.StatusOK, resp)
}
