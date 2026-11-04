package container

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/StasanDev/avitoLabs/internal/config"
	httptransport "github.com/StasanDev/avitoLabs/internal/http"
	"github.com/StasanDev/avitoLabs/internal/http/trip/create"
	"github.com/StasanDev/avitoLabs/internal/http/trip/finish"
	"github.com/StasanDev/avitoLabs/internal/http/trip/get"
	database "github.com/StasanDev/avitoLabs/internal/infra/postgres"
	repo "github.com/StasanDev/avitoLabs/internal/repository/postgres"
	createusecase "github.com/StasanDev/avitoLabs/internal/usecase/trip/create"
	finishusecase "github.com/StasanDev/avitoLabs/internal/usecase/trip/finish"
	getusecase "github.com/StasanDev/avitoLabs/internal/usecase/trip/get"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	pool   *pgxpool.Pool
	server *http.Server
}

func BuildApp(ctx context.Context, cfg *config.Config) (*App, error) {
	dbCtx, cancel := context.WithTimeout(ctx, cfg.DB.ConnectTimeout)
	defer cancel()
	pool, err := database.NewPostgresPool(dbCtx, &cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres pool: %w", err)
	}
	log.Printf("connected to db")

	tripRepository := repo.NewTripRepository(pool, cfg.DB.QueryTimeout)
	tripStatusHistoryRepository := repo.NewTripStatusRepository(pool, cfg.DB.QueryTimeout)
	idempotencyRepository := repo.NewIdempotencyRepository(pool, cfg.DB.QueryTimeout)
	txManager := repo.NewTxManager(pool, cfg.DB.QueryTimeout)
	createService := createusecase.NewService(
		tripRepository, txManager, idempotencyRepository,
		tripStatusHistoryRepository, cfg.IdempotencyTTL,
	)
	getService := getusecase.NewService(tripRepository)
	finishService := finishusecase.NewService(tripRepository, tripStatusHistoryRepository, txManager)
	httpHandler := httptransport.NewHandler(
		create.NewHandler(createService),
		get.NewHandler(getService),
		finish.NewHandler(finishService),
		pool,
		cfg.DB.QueryTimeout,
	)
	router, err := httptransport.NewRouter(httpHandler)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("build HTTP router: %w", err)
	}

	server := &http.Server{
		Addr:              cfg.Http.Addr,
		Handler:           router,
		ReadTimeout:       cfg.Http.ReadTimeout,
		ReadHeaderTimeout: cfg.Http.ReadHeaderTimeout,
		WriteTimeout:      cfg.Http.WriteTimeout,
		IdleTimeout:       cfg.Http.IdleTimeout,
	}

	return &App{
		pool:   pool,
		server: server,
	}, nil
}
