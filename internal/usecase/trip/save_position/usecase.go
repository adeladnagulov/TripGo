package createposition

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/adeladnagulov/TripGo/internal/http/helper/decode"
	"github.com/google/uuid"
)

type SavePositionUseCase struct {
	repo   positionSaver
	tm     txManager
	logger *slog.Logger
}

func NewSavePositionUseCase(repo positionSaver, tm txManager, logger *slog.Logger) *SavePositionUseCase {
	return &SavePositionUseCase{
		repo:   repo,
		tm:     tm,
		logger: logger,
	}
}

func (s *SavePositionUseCase) SavePosition(ctx context.Context, tripId uuid.UUID, posData trip.PositionData) (*trip.TripPosition, error) {
	pos := trip.TripPosition{
		TripId:      tripId,
		Coordinates: posData.Coordinates,
		RecordedAt:  posData.RecordedAt,
	}
	if pos.RecordedAt.After(time.Now().Add(time.Minute)) {
		return nil, decode.NewVlidationErr(trip.ErrRecordedAtInTheFuture)
	}
	err := s.tm.Do(ctx, func(ctx context.Context) error {
		posId, err := s.repo.SavePosition(ctx, pos)
		pos.Id = posId
		return err
	})
	if err != nil {
		s.logger.WarnContext(ctx, "save position error",
			slog.Any("error", err),
			"trip_id", tripId,
			"recorded_at", posData.RecordedAt,
		)
		return nil, fmt.Errorf("save position error: %w", err)
	}
	return &pos, nil
}
