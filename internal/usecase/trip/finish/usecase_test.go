package finish_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/adeladnagulov/TripGo/internal/usecase/trip/finish"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/metric/noop"
	"go.uber.org/mock/gomock"
)

func TestFinishTrip(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()

	tests := []struct {
		name       string
		setup      func(repo *MocktripFinisher, tm *MocktxManager)
		wantStatus trip.Status
		wantErr    error
	}{
		{
			name: "success",
			setup: func(repo *MocktripFinisher, tx *MocktxManager) {
				repo.EXPECT().
					GetByID(gomock.Any(), gomock.Any()).
					Return(&trip.Trip{
						Id:         uuid.New(),
						DriverId:   uuid.New(),
						UserId:     uuid.New(),
						StartPoint: trip.Coordinates{Lat: 12.12, Lon: 12.12},
						EndPoint:   trip.Coordinates{Lat: 13.13, Lon: 13.13},
						Price:      67,
						Status:     trip.Active,
						StartedAt:  time.Now(),
					}, nil)

				tx.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					})
				repo.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantStatus: trip.Completed,
			wantErr:    nil,
		},
		{
			name: "trip not found",
			setup: func(repo *MocktripFinisher, tx *MocktxManager) {
				repo.EXPECT().
					GetByID(gomock.Any(), gomock.Any()).
					Return(nil, trip.ErrNotFound)
			},
			wantErr: trip.ErrNotFound,
		},
		{
			name: "trip has already been completed",
			setup: func(repo *MocktripFinisher, tx *MocktxManager) {
				repo.EXPECT().
					GetByID(gomock.Any(), gomock.Any()).
					Return(&trip.Trip{
						Id:         uuid.New(),
						DriverId:   uuid.New(),
						UserId:     uuid.New(),
						StartPoint: trip.Coordinates{Lat: 12.12, Lon: 12.12},
						EndPoint:   trip.Coordinates{Lat: 13.13, Lon: 13.13},
						Price:      67,
						Status:     trip.Completed,
						StartedAt:  time.Now(),
					}, nil)
			},
			wantStatus: trip.Completed,
			wantErr:    trip.ErrTripCompleted,
		},
		{
			name: "failed save trip",
			setup: func(repo *MocktripFinisher, tx *MocktxManager) {
				repo.EXPECT().
					GetByID(gomock.Any(), gomock.Any()).
					Return(&trip.Trip{
						Id:         uuid.New(),
						DriverId:   uuid.New(),
						UserId:     uuid.New(),
						StartPoint: trip.Coordinates{Lat: 12.12, Lon: 12.12},
						EndPoint:   trip.Coordinates{Lat: 13.13, Lon: 13.13},
						Price:      67,
						Status:     trip.Active,
						StartedAt:  time.Now(),
					}, nil)

				tx.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					})
				repo.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Return(trip.ErrNotFound)
			},
			wantErr: trip.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMocktripFinisher(ctrl)
			mockTx := NewMocktxManager(ctrl)
			tt.setup(mockRepo, mockTx)
			us := finish.NewFinishUseCase(mockRepo, mockTx, slog.Default(), noop.NewMeterProvider().Meter("test"))

			result, err := us.FinishTrip(ctx, uuid.New())
			if err != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantStatus, result.Status)
			}
		})
	}
}
