-- name: CreateMovie :one
INSERT INTO movies (
        title,
        "year",
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