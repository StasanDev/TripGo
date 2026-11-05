package trip

import (
	"time"

	"github.com/google/uuid"
)

type TripStatus string

const (
	TripStatusActive    TripStatus = "active"
	TripStatusCompleted TripStatus = "completed"
)

type TripInput struct {
	UserID         uuid.UUID
	DriverID       uuid.UUID
	StartPoint     Coordinates
	EndPoint       Coordinates
	Price          int64
	IdempotencyKey *uuid.UUID
	RequestHash    []byte
}

type Trip struct {
	TripInput
	ID         uuid.UUID
	Status     TripStatus
	StartedAt  time.Time
	FinishedAt *time.Time
}

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

func NewTrip(input TripInput, id uuid.UUID, startedAt time.Time) (Trip, error) {
	if err := input.Validate(); err != nil {
		return Trip{}, err
	}

	return Trip{
		TripInput:  input,
		ID:         id,
		Status:     TripStatusActive,
		StartedAt:  startedAt,
		FinishedAt: nil,
	}, nil
}
