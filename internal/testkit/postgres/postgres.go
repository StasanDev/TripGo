//go:build integration

package testkit

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/StasanDev/avitoLabs/internal/infra/postgres/transaction/txmanager"
	"github.com/StasanDev/avitoLabs/migrations"
)

type Infra struct {
	Pool      *pgxpool.Pool
	TxManager *txmanager.TxManager
}

func Start(t *testing.T) *Infra {
	t.Helper()

	ctx := context.Background()

	container, err := testcontainers.Run(
		ctx,
		"postgres:16-alpine",
		testcontainers.WithEnv(map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_DB":       "testdb",
		}),
		testcontainers.WithExposedPorts("5432/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(2*time.Minute),
		),
	)
	require.NoError(t, err)

	testcontainers.CleanupContainer(t, container)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)

	address := net.JoinHostPort(host, port.Port())
	dsn := fmt.Sprintf(
		"postgres://postgres:password@%s/testdb?sslmode=disable",
		address,
	)

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)

	t.Cleanup(pool.Close)

	require.NoError(t, pool.Ping(ctx))
	applyMigrations(t, ctx, pool)

	return &Infra{
		Pool:      pool,
		TxManager: txmanager.NewTxManager(pool, 3*time.Second),
	}
}

func applyMigrations(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
) {
	t.Helper()

	db := stdlib.OpenDBFromPool(pool)
	defer func() {
		require.NoError(t, db.Close())
	}()

	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		db,
		migrations.FS,
	)
	require.NoError(t, err)

	_, err = provider.Up(ctx)
	require.NoError(t, err)
}

func (i *Infra) Reset(
	t *testing.T,
	ctx context.Context,
) {
	t.Helper()

	_, err := i.Pool.Exec(
		ctx,
		`
			TRUNCATE TABLE
				trip_idempotency_keys,
				trip_status_history,
				trips
			RESTART IDENTITY CASCADE
		`,
	)
	require.NoError(t, err)
}
