package trip

import (
	"time"

	"github.com/google/uuid"
)

type TripStatusHistory struct {
	TripID     uuid.UUID
	FromStatus *TripStatus
	ToStatus   TripStatus
	Reason     *string
	ChangedAt  time.Time
}


func NewTripStatusHistory(
	tripModel Trip,
	fromStatus *TripStatus,
	reason string,
	changedAt time.Time,
) TripStatusHistory {
	return TripStatusHistory{
		TripID:     tripModel.ID,
		FromStatus: fromStatus,
		ToStatus:   tripModel.Status,
		Reason:     &reason,
		ChangedAt:  changedAt,
	}
}
