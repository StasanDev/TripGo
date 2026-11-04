package create

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
	"github.com/google/uuid"
)

func (s *Service) CreateTrip(ctx context.Context, tripInp trip.TripInput) (trip.Trip, bool, error) {
	if err := tripInp.Validate(); err != nil {
		return trip.Trip{}, false, fmt.Errorf("%w: %w", trip.ErrInvalidInput, err)
	}

	now := time.Now().UTC()
	tripOut := trip.Trip{
		TripInput:  tripInp,
		ID:         uuid.New(),
		Status:     trip.TripStatusActive,
		StartedAt:  now,
		FinishedAt: nil,
	}

	reason := "trip created"
	history := trip.TripStatusHistory{
		TripID:    tripOut.ID,
		ToStatus:  trip.TripStatusActive,
		Reason:    &reason,
		ChangedAt: now,
	}

	resultTrip := tripOut
	replayed := false
	err := s.txManager.Do(ctx, func(txCtx context.Context) error {
		if tripInp.IdempotencyKey != nil {
			idempotencyRecord := trip.IdempotencyRecord{
				Key:         *tripInp.IdempotencyKey,
				RequestHash: tripInp.RequestHash,
				TripID:      tripOut.ID,
				ExpiresAt:   now.Add(s.idempotencyTTL),
			}

			existingRecord, reserved, err := s.idempotencyRepository.ReserveOrGet(txCtx, idempotencyRecord)
			if err != nil {
				return fmt.Errorf("reserve or get idempotency key: %w", err)
			}
			if !reserved {
				if !bytes.Equal(existingRecord.RequestHash, tripInp.RequestHash) {
					return trip.ErrIdempotencyConflict
				}

				resultTrip, err = s.tripRepository.GetTrip(txCtx, existingRecord.TripID)
				if err != nil {
					return fmt.Errorf("get trip by idempotency key: %w", err)
				}
				replayed = true
				return nil
			}
		}
		if err := s.tripRepository.CreateTrip(txCtx, tripOut); err != nil {
			return fmt.Errorf("create trip: %w", err)
		}
		if err := s.tripStatusHistoryRepository.AddInTripHistory(txCtx, history); err != nil {
			return fmt.Errorf("create initial trip status history: %w", err)
		}
		return nil
	})
	if err != nil {
		return trip.Trip{}, false, err
	}

	return resultTrip, replayed, nil
}
