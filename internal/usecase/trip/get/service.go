package get

type Service struct {
	tripRepository tripRepository
}

func NewService(tripRepository tripRepository) *Service {
	return &Service{
		tripRepository: tripRepository,
	}
}
