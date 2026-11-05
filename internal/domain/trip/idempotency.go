package trip

import (
	"time"

	"github.com/google/uuid"
)

type IdempotencyRecord struct {
	Key         uuid.UUID
	RequestHash []byte
	TripID      uuid.UUID
	ExpiresAt   time.Time
}

func NewIdempotencyRecord(
	key uuid.UUID,
	requestHash []byte,
	tripID uuid.UUID,
	expiresAt time.Time,
) IdempotencyRecord {
	return IdempotencyRecord{
		Key:         key,
		RequestHash: requestHash,
		TripID:      tripID,
		ExpiresAt:   expiresAt,
	}
}
