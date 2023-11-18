-- name: GetMovieBy :one
SELECT *
FROM movies
WHERE id = $1;