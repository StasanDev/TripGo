//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test

package create

import (
	"context"

	"github.com/google/uuid"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
)

type (
	tripRepository interface {
		CreateTrip(ctx context.Context, trip trip.Trip) error
		GetTrip(ctx context.Context, id uuid.UUID) (trip.Trip, error)
	}

	tripStatusHistoryRepository interface {
		AddInTripHistory(ctx context.Context, history trip.TripStatusHistory) error
	}

	txManager interface {
		Do(ctx context.Context, fn func(ctx context.Context) error) error
	}

	idempotencyRepository interface {
		ReserveOrGet(
			ctx context.Context,
			record trip.IdempotencyRecord,
		) (existing trip.IdempotencyRecord, reserved bool, err error)
	}
)
