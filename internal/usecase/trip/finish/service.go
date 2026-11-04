package finish

type Service struct {
	tripRepository              tripRepository
	tripStatusHistoryRepository tripStatusHistoryRepository
	txManager                   txManager
}

func NewService(
	tripRepository tripRepository,
	tripStatusHistoryRepository tripStatusHistoryRepository,
	txManager txManager,
) *Service {
	return &Service{
		tripRepository:              tripRepository,
		tripStatusHistoryRepository: tripStatusHistoryRepository,
		txManager:                   txManager,
	}
}
