package finish

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/adeladnagulov/TripGo/api"
	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/adeladnagulov/TripGo/internal/http/domainerror"
	"github.com/adeladnagulov/TripGo/internal/http/response"
	"github.com/adeladnagulov/TripGo/internal/usecase/trip/finish"
)

type Handler struct {
	uc *finish.FinishUseCase
}

func NewHandler(us *finish.FinishUseCase) *Handler {
	return &Handler{uc: us}
}

// FinishTrip Завершить поездку
// (POST /api/v1/trips/{tripId}/finish)
func (h *Handler) FinishTrip(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	t, err := h.uc.FinishTrip(r.Context(), tripId)
	if err != nil {
		if errors.Is(err, trip.ErrCannotFinishTrip) {
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
