package createposition

import (
	"net/http"

	"github.com/adeladnagulov/TripGo/api"
	"github.com/adeladnagulov/TripGo/internal/http/domainerror"
	"github.com/adeladnagulov/TripGo/internal/http/helper/decode"
	"github.com/adeladnagulov/TripGo/internal/http/helper/response"
)

type Handler struct {
	uc tripUseCase
}

func NewHandler(uc tripUseCase) *Handler {
	return &Handler{uc: uc}
}

// CreateTripPosition Сохранить координату поездки
// (POST /api/v1/trips/{tripId}/positions)
func (h *Handler) CreateTripPosition(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	var apiPosData api.PositionData
	if err := decode.DecodeJSON(r, &apiPosData); err != nil {
		status, problem := domainerror.Problem(err, r)
		response.Problem(w, status, problem)
		return
	}
	posData := toDomain(apiPosData)
	tripPos, err := h.uc.SavePosition(r.Context(), tripId, posData)
	if err != nil {
		status, problem := domainerror.Problem(err, r)
		response.Problem(w, status, problem)
		return
	}
	resp := toApi(*tripPos)
	response.Json(w, http.StatusCreated, resp)
}
