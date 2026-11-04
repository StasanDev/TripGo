package finish

import (
	"context"
	"fmt"
	"time"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
	"github.com/google/uuid"
)

func (s *Service) FinishTrip(ctx context.Context, id uuid.UUID) (trip.Trip, error) {
	finishAt := time.Now().UTC()
	fromStatus := trip.TripStatusActive
	reason := "trip finished"

	history := trip.TripStatusHistory{
		TripID:     id,
		FromStatus: &fromStatus,
		ToStatus:   trip.TripStatusCompleted,
		Reason:     &reason,
		ChangedAt:  finishAt,
	}

	var tripOut trip.Trip
	err := s.txManager.Do(ctx, func(txCtx context.Context) error {
		var err error
		tripOut, err = s.tripRepository.FinishTrip(txCtx, id, finishAt)
		if err != nil {
			return fmt.Errorf("finish trip: %w", err)
		}
		if err := s.tripStatusHistoryRepository.AddInTripHistory(txCtx, history); err != nil {
			return fmt.Errorf("add finished trip status history: %w", err)
		}
		return nil
	})

	if err != nil {
		return trip.Trip{}, err
	}

	return tripOut, nil
}
