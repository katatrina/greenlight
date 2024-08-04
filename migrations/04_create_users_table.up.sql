CREATE TABLE users (
    id bigserial PRIMARY KEY,
    name text NOT NULL,
    email citext UNIQUE NOT NULL,
    hashed_password bytea NOT NULL,
    activated bool NOT NULL,
    version integer NOT NULL DEFAULT 1,
    created_at timestamptz(0) NOT NULL DEFAULT NOW()
);