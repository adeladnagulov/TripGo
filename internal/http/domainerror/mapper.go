package domainerror

import (
	"errors"
	"net/http"

	"github.com/adeladnagulov/TripGo/api"
	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/adeladnagulov/TripGo/internal/http/helper/decode"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
)

type contextKey string

const RoutePatternKey contextKey = "openapi_route_pattern"

func Problem(err error, r *http.Request) (int, api.Problem) {
	if err == nil {
		return http.StatusOK, api.Problem{}
	}
	host := getRoutePattern(r)
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
	if err == nil {
		return false
	}
	var validateErr *decode.ValidationErr
	if errors.As(err, &validateErr) {
		return true
	}
	var reqErr *openapi3filter.RequestError
	if errors.As(err, &reqErr) {
		return true
	}
	var paramErr *api.InvalidParamFormatError
	if errors.As(err, &paramErr) {
		return true
	}

	return false
}

func getRoutePattern(r *http.Request) string {
	if pattern, ok := r.Context().Value(RoutePatternKey).(string); ok && pattern != "" {
		return pattern
	}

	if routeCtx := chi.RouteContext(r.Context()); routeCtx != nil {
		return routeCtx.RoutePattern()
	}

	return r.URL.Path
}
