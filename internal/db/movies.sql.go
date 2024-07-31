// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: movies.sql

package db

import (
	"context"
)

const createMovie = `-- name: CreateMovie :one
INSERT INTO movies (
        title,
        year,
        runtime,
        genres
    )
VALUES (
        $1,
        $2,
        $3,
        $4
    )
RETURNING id, title, runtime, genres, year, version, created_at
`

type CreateMovieParams struct {
	Title   string   `json:"title"`
	Year    int32    `json:"year"`
	Runtime Runtime  `json:"runtime"`
	Genres  []string `json:"genres"`
}

func (q *Queries) CreateMovie(ctx context.Context, arg CreateMovieParams) (Movie, error) {
	row := q.db.QueryRow(ctx, createMovie,
		arg.Title,
		arg.Year,
		arg.Runtime,
		arg.Genres,
	)
	var i Movie
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Runtime,
		&i.Genres,
		&i.Year,
		&i.Version,
		&i.CreatedAt,
	)
	return i, err
}

const deleteMovie = `-- name: DeleteMovie :execrows
DELETE FROM movies
WHERE id = $1
`

func (q *Queries) DeleteMovie(ctx context.Context, id int64) (int64, error) {
	result, err := q.db.Exec(ctx, deleteMovie, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const getMovie = `-- name: GetMovie :one
SELECT id, title, runtime, genres, year, version, created_at
FROM movies
WHERE id = $1
`

func (q *Queries) GetMovie(ctx context.Context, id int64) (Movie, error) {
	row := q.db.QueryRow(ctx, getMovie, id)
	var i Movie
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Runtime,
		&i.Genres,
		&i.Year,
		&i.Version,
		&i.CreatedAt,
	)
	return i, err
}

const updateMovie = `-- name: UpdateMovie :one
UPDATE movies
SET
    title = $2,
    year = $3,
    runtime = $4,
    genres = $5,
    version = version + 1
WHERE id = $1
RETURNING id, title, runtime, genres, year, version, created_at
`

type UpdateMovieParams struct {
	ID      int64    `json:"id"`
	Title   string   `json:"title"`
	Year    int32    `json:"year"`
	Runtime Runtime  `json:"runtime"`
	Genres  []string `json:"genres"`
}

func (q *Queries) UpdateMovie(ctx context.Context, arg UpdateMovieParams) (Movie, error) {
	row := q.db.QueryRow(ctx, updateMovie,
		arg.ID,
		arg.Title,
		arg.Year,
		arg.Runtime,
		arg.Genres,
	)
	var i Movie
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Runtime,
		&i.Genres,
		&i.Year,
		&i.Version,
		&i.CreatedAt,
	)
	return i, err
}
