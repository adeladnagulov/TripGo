package finish_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adeladnagulov/TripGo/api"
	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/adeladnagulov/TripGo/internal/http/trip/finish"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

func TestFinishTrip(t *testing.T) {
	tripId := uuid.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name       string
		setup      func(uc *MocktripUseCase)
		wantStatus int
		wantCode   string
	}{
		{
			name: "success",
			setup: func(uc *MocktripUseCase) {
				uc.EXPECT().
					FinishTrip(gomock.Any(), gomock.Any()).
					Return(&trip.Trip{Id: tripId, Status: trip.Completed}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "trip completed",
			setup: func(uc *MocktripUseCase) {
				uc.EXPECT().
					FinishTrip(gomock.Any(), gomock.Any()).
					Return(nil, trip.ErrTripCompleted)
			},
			wantStatus: http.StatusConflict,
			wantCode:   "trip_completed",
		},
		{
			name: "internal error",
			setup: func(uc *MocktripUseCase) {
				uc.EXPECT().
					FinishTrip(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("dbError"))
			},
			wantStatus: http.StatusInternalServerError,
			wantCode:   "internal_error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUc := NewMocktripUseCase(ctrl)
			tt.setup(mockUc)
			handler := finish.NewHandler(mockUc)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/api/v1/trips/676767/finish", bytes.NewBufferString(""))
			r.Header.Set("Content-Type", "application/json")

			handler.FinishTrip(w, r, tripId)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantCode != "" {
				var problem api.Problem
				json.NewDecoder(w.Body).Decode(&problem)
				assert.Equal(t, tt.wantCode, problem.Code)
			}
		})
	}
}
