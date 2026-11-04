package get

import (
	"context"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
	"github.com/google/uuid"
)

func (s *Service) GetTrip(ctx context.Context, id uuid.UUID) (trip.Trip, error) {
	return s.tripRepository.GetTrip(ctx, id)
}
