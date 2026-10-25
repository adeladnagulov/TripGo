package finish

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/adeladnagulov/TripGo/api"
	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/adeladnagulov/TripGo/internal/http/domainerror"
	"github.com/adeladnagulov/TripGo/internal/http/response"
)

type Handler struct {
	uc tripUseCase
}

func NewHandler(us tripUseCase) *Handler {
	return &Handler{uc: us}
}

// FinishTrip Завершить поездку
// (POST /api/v1/trips/{tripId}/finish)
func (h *Handler) FinishTrip(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	t, err := h.uc.FinishTrip(r.Context(), tripId)
	if err != nil {
		if errors.Is(err, trip.ErrTripCompleted) {
			status, problem := domainerror.Problem(err, r)
			response.Problem(w, status, problem)
			return
		}
		fmt.Println(err) //лог
		status, problem := domainerror.Problem(err, r)
		response.Problem(w, status, problem)
		return
	}
	resp := toApi(*t)
	response.Json(w, http.StatusOK, resp)
}
