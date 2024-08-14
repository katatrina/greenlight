package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	Querier
	GenerateToken(ctx context.Context, userID int64, ttl time.Duration, scope string) (tokenPlaintext string, err error)
}

type SQLStore struct {
	*Queries
}

func NewStore(connPool *pgxpool.Pool) Store {
	return &SQLStore{
		Queries: New(connPool),
	}
}
