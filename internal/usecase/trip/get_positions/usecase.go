package getpositions

import (
	"context"
	"log/slog"

	"github.com/adeladnagulov/TripGo/internal/domain/trip"
	"github.com/google/uuid"
)

type GetPosUseCase struct {
	repo   positionsGetter
	logger *slog.Logger
}

func NewGetUseCase(repo positionsGetter, logger *slog.Logger) *GetPosUseCase {
	return &GetPosUseCase{
		repo:   repo,
		logger: logger,
	}
}

func (s *GetPosUseCase) GetPositions(ctx context.Context, tripId uuid.UUID) (*trip.TripPositionsResponse, error) {
	positions, err := s.repo.GetPositions(ctx, tripId)
	if err != nil {
		s.logger.WarnContext(ctx, "get positions error",
			slog.String("error", err.Error()),
			"tripId", tripId,
		)
	}
	resp := trip.TripPositionsResponse{
		Positions: positions,
	}
	s.logger.InfoContext(ctx, "get positions successfully",
		"trip_id", tripId,
		"posCount", len(resp.Positions),
	)
	return &resp, nil
}
