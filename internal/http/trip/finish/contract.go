//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test

package finish

import (
	"context"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
	"github.com/google/uuid"
)

type tripFinisher interface {
	FinishTrip(ctx context.Context, id uuid.UUID) (trip.Trip, error)
}
