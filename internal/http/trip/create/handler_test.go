package create_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adeladnagulov/TripGo/api"
	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/adeladnagulov/TripGo/internal/http/trip/create"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

func TestCreateTrip(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name       string
		body       string
		setup      func(uc *MocktripUseCase)
		wantStatus int
		wantCode   string
	}{
		{
			name: "success",
			body: `{
  "user_id": "5cb72c04-7650-45c9-a79b-bcdba0631e0c",
  "driver_id": "8860b315-ec86-42eb-a17c-7c163d721ff5",
  "start_point": { "latitude": 59.9398, "longitude": 30.3146 },
  "end_point":   { "latitude": 59.9290, "longitude": 30.3626 },
  "price": 1450
}`,
			setup: func(uc *MocktripUseCase) {
				uc.EXPECT().
					CreateTrip(gomock.Any(), gomock.Any()).
					Return(&trip.Trip{Id: uuid.New(), Status: trip.Active}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid JSON",
			body:       `{invalid JSON}`,
			setup:      func(uc *MocktripUseCase) {},
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_request",
		},
		{
			name: "driver busy",
			body: `{
  "user_id": "5cb72c04-7650-45c9-a79b-bcdba0631e0c",
  "driver_id": "8860b315-ec86-42eb-a17c-7c163d721ff5",
  "start_point": { "latitude": 59.9398, "longitude": 30.3146 },
  "end_point":   { "latitude": 59.9290, "longitude": 30.3626 },
  "price": 1450
}`,
			setup: func(uc *MocktripUseCase) {
				uc.EXPECT().
					CreateTrip(gomock.Any(), gomock.Any()).
					Return(nil, trip.ErrDriverBusy)
			},
			wantStatus: http.StatusConflict,
			wantCode:   "driver_busy",
		},
		{
			name: "internal error",
			body: `{
  "user_id": "5cb72c04-7650-45c9-a79b-bcdba0631e0c",
  "driver_id": "8860b315-ec86-42eb-a17c-7c163d721ff5",
  "start_point": { "latitude": 59.9398, "longitude": 30.3146 },
  "end_point":   { "latitude": 59.9290, "longitude": 30.3626 },
  "price": 1450
}`,
			setup: func(uc *MocktripUseCase) {
				uc.EXPECT().
					CreateTrip(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantCode:   "internal_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUc := NewMocktripUseCase(ctrl)
			tt.setup(mockUc)
			handler := create.NewHandler(mockUc)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/api/v1/trips", bytes.NewBufferString(tt.body))
			r.Header.Set("Content-Type", "application/json")

			handler.CreateTrip(w, r, api.CreateTripParams{})

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantCode != "" {
				var problem api.Problem
				json.NewDecoder(w.Body).Decode(&problem)
				assert.Equal(t, tt.wantCode, problem.Code)
			}
		})
	}
}
