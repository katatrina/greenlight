-- name: CreateMovie :one
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
RETURNING *;

-- name: GetMovie :one
SELECT *
FROM movies
WHERE id = $1;

-- name: UpdateMovie :one
UPDATE movies
SET
    title = coalesce(sqlc.narg('title'), title),
    year = coalesce(sqlc.narg('year'), year),
    runtime = coalesce(sqlc.narg('runtime')::int, runtime),
    genres = coalesce(sqlc.narg('genres'), genres),
    version = version + 1
WHERE id = sqlc.arg('id') AND version = sqlc.arg('version')
RETURNING *;

-- name: DeleteMovie :execrows
DELETE FROM movies
WHERE id = $1;