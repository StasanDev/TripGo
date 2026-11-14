//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	repository "github.com/StasanDev/avitoLabs/internal/repository/postgres/trip"
	testkit "github.com/StasanDev/avitoLabs/internal/testkit/postgres"
)

type RepositorySuite struct {
	suite.Suite

	ctx                         context.Context
	infra                       *testkit.Infra
	tripRepo                    *repository.TripRepository
	tripStatusHistoryRepository *repository.TripStatusHistoryRepository
	idempotencyRepository       *repository.IdempotencyRepository
}

func (s *RepositorySuite) SetupSuite() {
	s.ctx = context.Background()
	s.infra = testkit.Start(s.T())

	timeout := 3 * time.Second

	s.tripRepo = repository.NewTripRepository(s.infra.Pool, timeout)
	s.tripStatusHistoryRepository = repository.NewTripStatusRepository(s.infra.Pool, timeout)
	s.idempotencyRepository = repository.NewIdempotencyRepository(s.infra.Pool, timeout)
}

func (s *RepositorySuite) SetupTest() {
	s.infra.Reset(s.T(), s.ctx)
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositorySuite))
}
