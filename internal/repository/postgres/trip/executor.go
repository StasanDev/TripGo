package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/StasanDev/avitoLabs/internal/infra/postgres/transaction/txkey"
)

func executor(ctx context.Context, pool *pgxpool.Pool) dbtx {
	if tx, ok := txkey.FromContext(ctx); ok {
		return tx
	}

	return pool
}
