//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test

package create

import (
	"context"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
)

type tripCreator interface {
	CreateTrip(ctx context.Context, input trip.TripInput) (trip.Trip, bool, error)
}
