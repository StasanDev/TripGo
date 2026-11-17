package trip_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
)

func TestTripInputValidate(t *testing.T) {
	tests := []struct {
		name        string
		changeField func(*trip.TripInput)
		wantField   string
		wantReason  string
	}{
		{
			name:        "valid input",
			changeField: func(*trip.TripInput) {},
		},
		{
			name: "boundary coordinates and zero price",
			changeField: func(input *trip.TripInput) {
				input.StartPoint = trip.Coordinates{Latitude: -90, Longitude: -180}
				input.EndPoint = trip.Coordinates{Latitude: 90, Longitude: 180}
				input.Price = 0
			},
		},
		{
			name: "empty user id",
			changeField: func(input *trip.TripInput) {
				input.UserID = uuid.Nil
			},
			wantField:  "user_id",
			wantReason: "must be a non-empty UUID",
		},
		{
			name: "empty driver id",
			changeField: func(input *trip.TripInput) {
				input.DriverID = uuid.Nil
			},
			wantField:  "driver_id",
			wantReason: "must be a non-empty UUID",
		},
		{
			name: "start latitude below minimum",
			changeField: func(input *trip.TripInput) {
				input.StartPoint.Latitude = -90.1
			},
			wantField:  "start_point.latitude",
			wantReason: "must be between -90 and 90",
		},
		{
			name: "start longitude above maximum",
			changeField: func(input *trip.TripInput) {
				input.StartPoint.Longitude = 180.1
			},
			wantField:  "start_point.longitude",
			wantReason: "must be between -180 and 180",
		},
		{
			name: "end latitude above maximum",
			changeField: func(input *trip.TripInput) {
				input.EndPoint.Latitude = 90.1
			},
			wantField:  "end_point.latitude",
			wantReason: "must be between -90 and 90",
		},
		{
			name: "end longitude below minimum",
			changeField: func(input *trip.TripInput) {
				input.EndPoint.Longitude = -180.1
			},
			wantField:  "end_point.longitude",
			wantReason: "must be between -180 and 180",
		},
		{
			name: "negative price",
			changeField: func(input *trip.TripInput) {
				input.Price = -1
			},
			wantField:  "price",
			wantReason: "must be greater than or equal to zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := validTripInput()
			tt.changeField(&input)

			err := input.Validate()
			if tt.wantField == "" {
				require.NoError(t, err)
				return
			}

			var validationErr *trip.ValidationError
			require.ErrorAs(t, err, &validationErr)
			assert.Equal(t, tt.wantField, validationErr.Field)
			assert.Equal(t, tt.wantReason, validationErr.Reason)
			assert.Equal(t, tt.wantField+": "+tt.wantReason, validationErr.Error())
		})
	}
}

func validTripInput() trip.TripInput {
	return trip.TripInput{
		UserID:   uuid.New(),
		DriverID: uuid.New(),
		StartPoint: trip.Coordinates{
			Latitude:  55.7558,
			Longitude: 37.6173,
		},
		EndPoint: trip.Coordinates{
			Latitude:  59.9343,
			Longitude: 30.3351,
		},
		Price: 1000,
	}
}
