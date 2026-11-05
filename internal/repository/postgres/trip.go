package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/StasanDev/avitoLabs/internal/domain/trip"
	"github.com/StasanDev/avitoLabs/internal/infra/postgres/transaction/txkey"
)

var queryTripCols = []string{
	"id", "user_id", "driver_id", "start_latitude", "start_longitude",
	"end_latitude", "end_longitude", "price", tripStatusColumn,
	"started_at", "finished_at",
}

const (
	tripsTable              = "trips"
	tripStatusHistoryTable  = "trip_status_history"
	driverActiveUniqueIndex = "trips_driver_active_uidx"
	tripStatusColumn        = "status"
)

type TripRepository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

func NewTripRepository(pool *pgxpool.Pool, queryTimeout time.Duration) *TripRepository {
	return &TripRepository{
		pool:         pool,
		queryTimeout: queryTimeout,
	}
}

func (r *TripRepository) CreateTrip(ctx context.Context, tripInp trip.Trip) error {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query, args, err := sq.
		Insert(tripsTable).Columns(queryTripCols...).
		Values(
			tripInp.ID, tripInp.UserID, tripInp.DriverID,
			tripInp.StartPoint.Latitude, tripInp.StartPoint.Longitude,
			tripInp.EndPoint.Latitude, tripInp.EndPoint.Longitude,
			tripInp.Price, tripInp.Status, tripInp.StartedAt, tripInp.FinishedAt,
		).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("build create trip: %w", err)
	}

	if _, err := r.executor(ctx).Exec(ctx, query, args...); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == driverActiveUniqueIndex {
			return trip.ErrDriverBusy
		}
		return fmt.Errorf("insert trip: %w", err)
	}

	return nil
}

func (r *TripRepository) GetTrip(ctx context.Context, id uuid.UUID) (trip.Trip, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query, args, err := sq.Select(queryTripCols...).From(tripsTable).Where(sq.Eq{"id": id}).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return trip.Trip{}, fmt.Errorf("build get trip: %w", err)
	}

	row := r.executor(ctx).QueryRow(ctx, query, args...)
	tripOut, err := scanTrip(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return trip.Trip{}, trip.ErrTripNotFound
		}
		return trip.Trip{}, fmt.Errorf("get trip: %w", err)
	}
	return tripOut, nil
}

func (r *TripRepository) FinishTrip(ctx context.Context, id uuid.UUID, at time.Time) (trip.Trip, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	query, args, err := sq.Update(tripsTable).SetMap(map[string]interface{}{
		tripStatusColumn: trip.TripStatusCompleted,
		"finished_at":    at,
		"updated_at":     at,
	}).Where(sq.Eq{
		"id":             id,
		tripStatusColumn: trip.TripStatusActive,
	}).Suffix("RETURNING " + strings.Join(queryTripCols, ", ")).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return trip.Trip{}, fmt.Errorf("build finish trip: %w", err)
	}

	row := r.executor(ctx).QueryRow(ctx, query, args...)
	tripOut, err := scanTrip(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_, err := r.GetTrip(ctx, id)
			if errors.Is(err, trip.ErrTripNotFound) {
				return trip.Trip{}, trip.ErrTripNotFound
			}
			if err != nil {
				return trip.Trip{}, fmt.Errorf("check trip after unsuccessful finish: %w", err)
			}
			return trip.Trip{}, trip.ErrTripCompleted
		}
		return trip.Trip{}, fmt.Errorf("finish trip: %w", err)
	}
	return tripOut, nil
}

func (r *TripRepository) executor(ctx context.Context) dbtx {
	if tx, ok := txkey.FromContext(ctx); ok {
		return tx
	}
	return r.pool
}

func scanTrip(row pgx.Row) (trip.Trip, error) {
	var tripOut trip.Trip
	err := row.Scan(&tripOut.ID, &tripOut.UserID, &tripOut.DriverID,
		&tripOut.StartPoint.Latitude, &tripOut.StartPoint.Longitude,
		&tripOut.EndPoint.Latitude, &tripOut.EndPoint.Longitude,
		&tripOut.Price, &tripOut.Status, &tripOut.StartedAt, &tripOut.FinishedAt)
	if err != nil {
		return trip.Trip{}, fmt.Errorf("scan trip: %w", err)
	}
	return tripOut, nil
}
