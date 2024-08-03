-- name: CreateMovie :one
INSERT INTO movies ( title, publish_year, runtime, genres)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetMovie :one
SELECT *
FROM movies
WHERE id = $1;

-- name: UpdateMovie :one
UPDATE movies
SET
    title = coalesce(sqlc.narg('title'), title),
    publish_year = coalesce(sqlc.narg('publish_year'), publish_year),
    runtime = coalesce(sqlc.narg('runtime')::int, runtime),
    genres = coalesce(sqlc.narg('genres'), genres),
    version = version + 1
WHERE id = sqlc.arg('id') AND version = sqlc.arg('version')
RETURNING *;

-- name: DeleteMovie :execrows
DELETE FROM movies
WHERE id = $1;

-- name: ListMoviesWithFilters :many
SELECT count(*) OVER() as total_records, sqlc.embed(movies)
FROM movies
WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', sqlc.arg('title')) OR sqlc.arg('title') = '')
AND (genres @> sqlc.arg('genres') OR sqlc.arg('genres') = '{}')
ORDER BY CASE
    WHEN NOT sqlc.arg('reverse')::boolean AND sqlc.arg('order_by')::text = 'id' THEN id
    WHEN NOT sqlc.arg('reverse')::boolean AND sqlc.arg('order_by')::text = 'publishYear' THEN publish_year
    WHEN NOT sqlc.arg('reverse')::boolean AND sqlc.arg('order_by')::text = 'runtime' THEN runtime
END ASC, CASE
    WHEN sqlc.arg('reverse')::boolean AND sqlc.arg('order_by')::text = 'id' THEN id
    WHEN sqlc.arg('reverse')::boolean AND sqlc.arg('order_by')::text = 'publishYear' THEN publish_year
    WHEN sqlc.arg('reverse')::boolean AND sqlc.arg('order_by')::text = 'runtime' THEN runtime
END  DESC, CASE
    WHEN NOT sqlc.arg('reverse')::boolean AND sqlc.arg('order_by')::text = 'title' THEN title
END ASC, CASE
    WHEN sqlc.arg('reverse')::boolean AND sqlc.arg('order_by')::text = 'title' THEN title
END DESC, id ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');