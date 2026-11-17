package trip

import "time"

func (t *Trip) Finish(at time.Time) error {
	if t.Status != TripStatusActive {
		return ErrTripCompleted
	}

	t.Status = TripStatusCompleted
	t.FinishedAt = &at
	return nil
}
