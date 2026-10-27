package domainerror

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/adeladnagulov/TripGo/api"
	"github.com/adeladnagulov/TripGo/internal/domain/trip"
)

func Problem(err error, r *http.Request) (int, api.Problem) {
	if err == nil {
		return http.StatusOK, api.Problem{}
	}
	host := r.URL.Host
	typeConflict := "about:blank"
	switch {
	case isValidationError(err):
		detail := "invalid request body or parameters"
		return http.StatusBadRequest, api.BadRequest{
			Type:     typeConflict,
			Title:    http.StatusText(http.StatusBadRequest),
			Status:   http.StatusBadRequest,
			Detail:   &detail,
			Code:     "invalid_request",
			Instance: &host,
		}
	case errors.Is(err, trip.ErrDriverBusy):
		detail := "driver already has active trip"
		return http.StatusConflict, api.CreateTripConflict{
			Type:     typeConflict,
			Title:    http.StatusText(http.StatusConflict),
			Status:   http.StatusConflict,
			Detail:   &detail,
			Code:     "driver_busy",
			Instance: &host,
		}
	case errors.Is(err, trip.ErrTripCompleted):
		detail := "trip is already over"
		return http.StatusConflict, api.BadRequest{
			Type:     typeConflict,
			Title:    http.StatusText(http.StatusConflict),
			Status:   http.StatusConflict,
			Detail:   &detail,
			Code:     "trip_completed",
			Instance: &host,
		}
	case errors.Is(err, trip.ErrNotFound):
		detail := "trip with an id does not exist"
		return http.StatusNotFound, api.BadRequest{
			Type:     typeConflict,
			Title:    http.StatusText(http.StatusNotFound),
			Status:   http.StatusNotFound,
			Detail:   &detail,
			Code:     "trip_not_found",
			Instance: &host,
		}
	}

	detail := "internal server error"
	return http.StatusInternalServerError, api.InternalError{
		Type:     typeConflict,
		Title:    http.StatusText(http.StatusInternalServerError),
		Status:   http.StatusInternalServerError,
		Detail:   &detail,
		Code:     "internal_error",
		Instance: &host,
	}
}

func isValidationError(err error) bool {
	var syntaxErr *json.SyntaxError
	var unmarshalErr *json.UnmarshalTypeError
	if errors.As(err, &syntaxErr) || errors.As(err, &unmarshalErr) {
		return true
	}
	return false
}
