package create

import "time"

type Service struct {
	tripRepository              tripRepository
	txManager                   txManager
	idempotencyRepository       idempotencyRepository
	tripStatusHistoryRepository tripStatusHistoryRepository
	idempotencyTTL              time.Duration
}

func NewService(
	tripRepository tripRepository,
	txManager txManager,
	idempotencyRepository idempotencyRepository,
	tripStatusHistoryRepository tripStatusHistoryRepository,
	idempotencyTTL time.Duration,
) *Service {
	return &Service{
		tripRepository:              tripRepository,
		txManager:                   txManager,
		idempotencyRepository:       idempotencyRepository,
		tripStatusHistoryRepository: tripStatusHistoryRepository,
		idempotencyTTL:              idempotencyTTL,
	}
}
