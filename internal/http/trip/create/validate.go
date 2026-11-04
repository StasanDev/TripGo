package create

import (
	"errors"
	"fmt"

	api "github.com/StasanDev/avitoLabs/internal/generated"
	"github.com/google/uuid"
)

func validateTripData(body api.TripData) error {
	if body.UserId == uuid.Nil {
		return errors.New("user_id must be a non-empty UUID")
	}
	if body.DriverId == uuid.Nil {
		return errors.New("driver_id must be a non-empty UUID")
	}
	if err := validateCoordinates("start_point", body.StartPoint); err != nil {
		return err
	}
	if err := validateCoordinates("end_point", body.EndPoint); err != nil {
		return err
	}
	if body.Price < 0 {
		return errors.New("price must be greater than or equal to zero")
	}
	return nil
}

func validateCoordinates(name string, coordinates api.Coordinates) error {
	if coordinates.Latitude < -90 || coordinates.Latitude > 90 {
		return fmt.Errorf("%s.latitude must be between -90 and 90", name)
	}
	if coordinates.Longitude < -180 || coordinates.Longitude > 180 {
		return fmt.Errorf("%s.longitude must be between -180 and 180", name)
	}
	return nil
}
