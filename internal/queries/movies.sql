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

-- name: ListMoviesWithFilters :many
SELECT *
FROM movies
WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', @title) OR @title = '')
AND (genres @> sqlc.arg(genres) OR sqlc.arg(genres) = '{}')
ORDER BY CASE
    WHEN NOT @reverse::boolean AND @order_by::text = 'id' THEN id
    WHEN NOT @reverse::boolean AND @order_by::text = 'year' THEN year
    WHEN NOT @reverse::boolean AND @order_by::text = 'runtime' THEN runtime
END ASC, CASE
    WHEN @reverse::boolean AND @order_by::text = 'id' THEN id
    WHEN @reverse::boolean AND @order_by::text = 'year' THEN year
    WHEN @reverse::boolean AND @order_by::text = 'runtime' THEN runtime
END  DESC, CASE
    WHEN NOT @reverse::boolean AND sqlc.arg(order_by)::text = 'title' THEN title
END ASC, CASE
    WHEN @reverse::boolean AND sqlc.arg(order_by)::text = 'title' THEN title
END DESC,
id ASC;