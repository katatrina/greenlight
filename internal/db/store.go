package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	Querier
	GenerateToken(ctx context.Context, arg GenerateTokenParams) (tokenPlaintext string, token Token, err error)
}

type SQLStore struct {
	*Queries
}

func NewStore(connPool *pgxpool.Pool) Store {
	return &SQLStore{
		Queries: New(connPool),
	}
}
