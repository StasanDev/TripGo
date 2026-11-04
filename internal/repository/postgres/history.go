package repository

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/StasanDev/avitoLabs/internal/domain/trip"
	"github.com/StasanDev/avitoLabs/internal/infra/postgres/transaction/txkey"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TripStatusHistoryRepository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

func NewTripStatusRepository(pool *pgxpool.Pool, queryTimeout time.Duration) *TripStatusHistoryRepository {
	return &TripStatusHistoryRepository{
		pool:         pool,
		queryTimeout: queryTimeout,
	}
}

func (r *TripStatusHistoryRepository) AddInTripHistory(ctx context.Context, history trip.TripStatusHistory) error {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query, args, err := sq.Insert(tripStatusHistoryTable).Columns("trip_id", "from_status", "to_status", "reason", "changed_at").
		Values(history.TripID, history.FromStatus, history.ToStatus,
			history.Reason, history.ChangedAt).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("build trip status history: %w", err)
	}

	if _, err := r.executor(ctx).Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("add in trip status history: %w", err)
	}

	return nil
}

func (r *TripStatusHistoryRepository) executor(ctx context.Context) DBTX {
	if tx, ok := txkey.FromContext(ctx); ok {
		return tx
	}
	return r.pool
}
