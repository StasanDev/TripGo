package trip

import (
	"errors"

	"github.com/google/uuid"
)

func (t *TripInput) Validate() error {
	if t.UserID == uuid.Nil {
		return errors.New("user_id must be a non-empty UUID")
	}
	if t.DriverID == uuid.Nil {
		return errors.New("driver_id must be a non-empty UUID")
	}
	if err := validateCoordinates("start_point", t.StartPoint); err != nil {
		return err
	}
	if err := validateCoordinates("end_point", t.EndPoint); err != nil {
		return err
	}
	if t.Price < 0 {
		return errors.New("price must be greater than or equal to zero")
	}
	return nil
}

func validateCoordinates(name string, coordinates Coordinates) error {
	if coordinates.Latitude < -90 || coordinates.Latitude > 90 {
		return errors.New(name + ".latitude must be between -90 and 90")
	}
	if coordinates.Longitude < -180 || coordinates.Longitude > 180 {
		return errors.New(name + ".longitude must be between -180 and 180")
	}
	return nil
}
