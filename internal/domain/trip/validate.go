package trip

import "github.com/google/uuid"

func (t *TripInput) Validate() error {
	if t.UserID == uuid.Nil {
		return &ValidationError{Field: "user_id", Reason: "must be a non-empty UUID"}
	}
	if t.DriverID == uuid.Nil {
		return &ValidationError{Field: "driver_id", Reason: "must be a non-empty UUID"}
	}
	if err := validateCoordinates("start_point", t.StartPoint); err != nil {
		return err
	}
	if err := validateCoordinates("end_point", t.EndPoint); err != nil {
		return err
	}
	if t.Price < 0 {
		return &ValidationError{Field: "price", Reason: "must be greater than or equal to zero"}
	}
	return nil
}

func validateCoordinates(name string, coordinates Coordinates) error {
	if coordinates.Latitude < -90 || coordinates.Latitude > 90 {
		return &ValidationError{Field: name + ".latitude", Reason: "must be between -90 and 90"}
	}
	if coordinates.Longitude < -180 || coordinates.Longitude > 180 {
		return &ValidationError{Field: name + ".longitude", Reason: "must be between -180 and 180"}
	}
	return nil
}
