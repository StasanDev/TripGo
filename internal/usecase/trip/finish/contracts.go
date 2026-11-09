//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test

package finish

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
)

type (
	tripRepository interface {
		GetTrip(ctx context.Context, id uuid.UUID) (trip.Trip, error)
		FinishTrip(ctx context.Context, id uuid.UUID, at time.Time) (trip.Trip, error)
	}

	tripStatusHistoryRepository interface {
		AddInTripHistory(ctx context.Context, history trip.TripStatusHistory) error
	}

	txManager interface {
		Do(ctx context.Context, fn func(ctx context.Context) error) error
	}

	recorder interface {
		Complete(ctx context.Context, tripID uuid.UUID)
	}
)
