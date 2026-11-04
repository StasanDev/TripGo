package finish

import (
	"context"
	"fmt"
	"time"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
	"github.com/google/uuid"
)

func (s *Service) FinishTrip(ctx context.Context, id uuid.UUID) (trip.Trip, error) {
	var tripOut trip.Trip
	err := s.txManager.Do(ctx, func(txCtx context.Context) error {
		tripOut, err := s.tripRepository.GetTrip(txCtx, id)
		if err != nil {
			return fmt.Errorf("get trip to finish: %w", err)
		}

		fromStatus := tripOut.Status
		if err := tripOut.Finish(time.Now().UTC()); err != nil {
			return fmt.Errorf("finish trip: %w", err)
		}

		tripOut, err = s.tripRepository.FinishTrip(txCtx, id, *tripOut.FinishedAt)
		if err != nil {
			return fmt.Errorf("finish trip: %w", err)
		}

		reason := "trip finished"
		history := trip.TripStatusHistory{
			TripID:     tripOut.ID,
			FromStatus: &fromStatus,
			ToStatus:   tripOut.Status,
			Reason:     &reason,
			ChangedAt:  *tripOut.FinishedAt,
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
