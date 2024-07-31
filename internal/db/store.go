package db

import "github.com/jackc/pgx/v5/pgxpool"

type Store interface {
	Querier
}

type SQLStore struct {
	*Queries
}

func NewStore(connPool *pgxpool.Pool) Store {
	return &SQLStore{
		Queries: New(connPool),
	}
}
