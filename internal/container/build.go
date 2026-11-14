package container

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/StasanDev/avitoLabs/internal/config"
	httptransport "github.com/StasanDev/avitoLabs/internal/http"
	"github.com/StasanDev/avitoLabs/internal/http/trip/create"
	"github.com/StasanDev/avitoLabs/internal/http/trip/finish"
	"github.com/StasanDev/avitoLabs/internal/http/trip/get"
	database "github.com/StasanDev/avitoLabs/internal/infra/postgres/pool"
	"github.com/StasanDev/avitoLabs/internal/infra/postgres/transaction/txmanager"
	"github.com/StasanDev/avitoLabs/internal/infra/telemetry"
	repo "github.com/StasanDev/avitoLabs/internal/repository/postgres/trip"
	createusecase "github.com/StasanDev/avitoLabs/internal/usecase/trip/create"
	finishusecase "github.com/StasanDev/avitoLabs/internal/usecase/trip/finish"
	getusecase "github.com/StasanDev/avitoLabs/internal/usecase/trip/get"
)

type App struct {
	pool               *pgxpool.Pool
	server             *http.Server
	telemetryProviders *telemetry.Providers
	shutdownTimeout    time.Duration
}

func BuildApp(
	ctx context.Context,
	cfg *config.Config,
	logger *slog.Logger,
	serviceVersion string,
) (*App, error) {
	providers, err := telemetry.NewProviders(
		ctx,
		cfg.OTel.ServiceName,
		serviceVersion,
		cfg.ShutdownTimeout,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("create telemetry providers: %w", err)
	}

	dbCtx, cancel := context.WithTimeout(ctx, cfg.DB.ConnectTimeout)
	defer cancel()
	pool, err := database.NewPostgresPool(
		dbCtx,
		&cfg.DB,
		providers.Tracer,
		providers.Meter,
	)
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("failed to create postgres pool: %w", err),
			shutdownTelemetryProviders(providers, cfg.ShutdownTimeout),
		)
	}
	logger.Info("connected to db")

	tripRepository := repo.NewTripRepository(pool, cfg.DB.QueryTimeout)
	tripStatusHistoryRepository := repo.NewTripStatusRepository(pool, cfg.DB.QueryTimeout)
	idempotencyRepository := repo.NewIdempotencyRepository(pool, cfg.DB.QueryTimeout)
	txManager := txmanager.NewTxManager(pool, cfg.DB.QueryTimeout)
	createService := createusecase.NewService(
		tripRepository, txManager, idempotencyRepository,
		tripStatusHistoryRepository, cfg.IdempotencyTTL,
		providers,
	)
	getService := getusecase.NewService(tripRepository)
	finishService := finishusecase.NewService(tripRepository, tripStatusHistoryRepository, txManager, providers)
	httpHandler := httptransport.NewHandler(
		create.NewHandler(createService, logger),
		get.NewHandler(getService, logger),
		finish.NewHandler(finishService, logger),
		pool,
		cfg.DB.QueryTimeout,
		logger,
	)
	router, err := httptransport.NewRouter(
		httpHandler,
		logger,
		providers.Tracer,
		providers.Meter,
	)
	if err != nil {
		pool.Close()
		return nil, errors.Join(
			fmt.Errorf("build HTTP router: %w", err),
			shutdownTelemetryProviders(providers, cfg.ShutdownTimeout),
		)
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
		pool:               pool,
		server:             server,
		telemetryProviders: providers,
		shutdownTimeout:    cfg.ShutdownTimeout,
	}, nil
}
