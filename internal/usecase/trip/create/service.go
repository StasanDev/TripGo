package create

import "time"

type Service struct {
	tripRepository              tripRepository
	txManager                   txManager
	idempotencyRepository       idempotencyRepository
	tripStatusHistoryRepository tripStatusHistoryRepository
	idempotencyTTL              time.Duration
	recorder                    recorder
}

func NewService(
	tripRepository tripRepository,
	txManager txManager,
	idempotencyRepository idempotencyRepository,
	tripStatusHistoryRepository tripStatusHistoryRepository,
	idempotencyTTL time.Duration,
	recorder recorder,
) *Service {
	return &Service{
		tripRepository:              tripRepository,
		txManager:                   txManager,
		idempotencyRepository:       idempotencyRepository,
		tripStatusHistoryRepository: tripStatusHistoryRepository,
		idempotencyTTL:              idempotencyTTL,
		recorder:                    recorder,
	}
}
