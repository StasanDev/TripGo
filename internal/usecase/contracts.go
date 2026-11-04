package usecase

import (
	"context"
	"time"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
	"github.com/google/uuid"
)

type tripRepository interface {
	CreateTrip(ctx context.Context, trip trip.Trip) error
	GetTrip(ctx context.Context, id uuid.UUID) (trip.Trip, error)
	FinishTrip(ctx context.Context, id uuid.UUID, at time.Time) (trip.Trip, error)
}

type tripStatusHistoryRepository interface {
	AddInTripHistory(ctx context.Context, history trip.TripStatusHistory) error
}

type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type idempotencyRepository interface {
	ReserveOrGet(
		ctx context.Context,
		record trip.IdempotencyRecord,
	) (existing trip.IdempotencyRecord, reserved bool, err error)
}
