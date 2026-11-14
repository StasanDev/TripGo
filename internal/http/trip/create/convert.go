package create

import (
	"github.com/StasanDev/avitoLabs/internal/domain/trip"
	api "github.com/StasanDev/avitoLabs/internal/generated"
)

func tripInputFromAPI(body api.TripData) trip.TripInput {
	return trip.TripInput{
		UserID:   body.UserId,
		DriverID: body.DriverId,
		StartPoint: trip.Coordinates{
			Latitude:  body.StartPoint.Latitude,
			Longitude: body.StartPoint.Longitude,
		},
		EndPoint: trip.Coordinates{
			Latitude:  body.EndPoint.Latitude,
			Longitude: body.EndPoint.Longitude,
		},
		Price: body.Price,
	}
}
