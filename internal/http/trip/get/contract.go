//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test

package get

import (
	"context"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
	"github.com/google/uuid"
)

type tripGetter interface {
	GetTrip(ctx context.Context, id uuid.UUID) (trip.Trip, error)
}
