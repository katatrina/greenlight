-- name: GetMovie :one
SELECT *
FROM movies
WHERE id = $1;