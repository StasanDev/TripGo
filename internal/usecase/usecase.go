package usecase

import (
	"github.com/StasanDev/avitoLabs/internal/usecase/trip/create"
	"github.com/StasanDev/avitoLabs/internal/usecase/trip/finish"
	"github.com/StasanDev/avitoLabs/internal/usecase/trip/get"
)

type TripService struct {
	create *create.Service
	get    *get.Service
	finish *finish.Service
}

func NewTripService(create *create.Service, get *get.Service, finish *finish.Service) *TripService {
	return &TripService{
		create: create,
		get:    get,
		finish: finish,
	}
}
