package db

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	ForeignKeyViolation = "23503"
	UniqueViolation     = "23505"
)

var (
	ErrRecordNotFound       = pgx.ErrNoRows
	ErrInvalidRuntimeFormat = errors.New("invalid runtime format")
)

// ErrorCode return the condition name of SQLSTATE error code returned by PostgreSQL server.
//
// See https://www.postgresql.org/docs/11/errcodes-appendix.html
func ErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}

	return ""
}

func IsContainErrorMessage(err error, value string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return strings.Contains(pgErr.Message, value)
	}

	return false
}
