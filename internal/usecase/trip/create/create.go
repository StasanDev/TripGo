package create

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
)

func (s *Service) CreateTrip(ctx context.Context, tripInp trip.TripInput) (trip.Trip, bool, error) {
	now := time.Now().UTC()
	tripOut, err := trip.NewTrip(tripInp, uuid.New(), now)
	if err != nil {
		return trip.Trip{}, false, fmt.Errorf("create trip: %w", err)
	}

	history := trip.NewTripStatusHistory(tripOut, nil, trip.TripCreatedReason, tripOut.StartedAt)

	resultTrip := tripOut
	replayed := false
	err = s.txManager.Do(ctx, func(txCtx context.Context) error {
		if tripInp.IdempotencyKey != nil {
			idempotencyRecord := trip.NewIdempotencyRecord(
				*tripInp.IdempotencyKey,
				tripInp.RequestHash,
				tripOut.ID,
				tripOut.StartedAt.Add(s.idempotencyTTL),
			)

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

	if !replayed && s.recorder != nil {
		s.recorder.Create(ctx, resultTrip.ID)
	}
	return resultTrip, replayed, nil
}
