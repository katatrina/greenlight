CREATE TABLE "users"
(
    "id"            bigserial PRIMARY KEY,
    "name"          text          NOT NULL,
    "email"         citext UNIQUE NOT NULL,
    "password_hash" bytea         NOT NULL,
    "activated"     bool          NOT NULL,
    "version"       integer       NOT NULL DEFAULT 1,
    "created_at"    timestamptz   NOT NULL DEFAULT (now())
);