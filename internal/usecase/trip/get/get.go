package get

import (
	"context"

	"github.com/google/uuid"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
)

func (s *Service) GetTrip(ctx context.Context, id uuid.UUID) (trip.Trip, error) {
	return s.tripRepository.GetTrip(ctx, id)
}
