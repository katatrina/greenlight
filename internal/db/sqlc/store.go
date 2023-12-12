package db

import "database/sql"

type Store struct {
	*Queries
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		Queries: New(db),
	}
}
