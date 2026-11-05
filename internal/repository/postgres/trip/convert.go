package repository

import (
	"fmt"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
	"github.com/jackc/pgx/v5"
)

func scanTrip(row pgx.Row) (trip.Trip, error) {
	var tripOut trip.Trip
	err := row.Scan(&tripOut.ID, &tripOut.UserID, &tripOut.DriverID,
		&tripOut.StartPoint.Latitude, &tripOut.StartPoint.Longitude,
		&tripOut.EndPoint.Latitude, &tripOut.EndPoint.Longitude,
		&tripOut.Price, &tripOut.Status, &tripOut.StartedAt, &tripOut.FinishedAt)
	if err != nil {
		return trip.Trip{}, fmt.Errorf("scan trip: %w", err)
	}
	return tripOut, nil
}
