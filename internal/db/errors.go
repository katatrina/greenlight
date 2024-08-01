package db

import (
	"errors"

	"github.com/jackc/pgx/v5"
)

var (
	ErrRecordNotFound       = pgx.ErrNoRows
	ErrInvalidRuntimeFormat = errors.New("invalid runtime format")
	ErrEditConflict         = errors.New("edit conflict")
)
