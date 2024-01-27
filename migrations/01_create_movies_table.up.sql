CREATE TABLE IF NOT EXISTS "movies"
(
    "id"         bigserial PRIMARY KEY,
    "title"      text        NOT NULL,
    "year"       int         NOT NULL,
    "runtime"    int         NOT NULL,
    "genres"     text[] NOT NULL,
    "version"    int         NOT NULL DEFAULT 1,
    "created_at" timestamptz NOT NULL DEFAULT (now())
);
