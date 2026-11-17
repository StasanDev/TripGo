package finish

type Service struct {
	tripRepository              tripRepository
	tripStatusHistoryRepository tripStatusHistoryRepository
	txManager                   txManager
	recorder                    recorder
}

func NewService(
	tripRepository tripRepository,
	tripStatusHistoryRepository tripStatusHistoryRepository,
	txManager txManager,
	recorder recorder,
) *Service {
	return &Service{
		tripRepository:              tripRepository,
		tripStatusHistoryRepository: tripStatusHistoryRepository,
		txManager:                   txManager,
		recorder:                    recorder,
	}
}
