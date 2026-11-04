package usecase

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
	"github.com/google/uuid"
)

type TripService struct {
	tripRepository              tripRepository
	tripStatusHistoryRepository tripStatusHistoryRepository
	idempotencyRepository       idempotencyRepository
	txManager                   TxManager
	idempotencyTTL              time.Duration
}

func NewTripService(
	tripRepository tripRepository,
	tripStatusHistoryRepository tripStatusHistoryRepository,
	idempotencyRepository idempotencyRepository,
	txManager TxManager,
	idempotencyTTL time.Duration,
) *TripService {
	return &TripService{
		tripRepository:              tripRepository,
		tripStatusHistoryRepository: tripStatusHistoryRepository,
		idempotencyRepository:       idempotencyRepository,
		txManager:                   txManager,
		idempotencyTTL:              idempotencyTTL,
	}
}

func (s *TripService) CreateTrip(ctx context.Context, tripInp trip.TripInput) (trip.Trip, bool, error) {
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

func (s *TripService) GetTrip(ctx context.Context, id uuid.UUID) (trip.Trip, error) {
	return s.tripRepository.GetTrip(ctx, id)
}

func (s *TripService) FinishTrip(ctx context.Context, id uuid.UUID) (trip.Trip, error) {
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
