CREATE TABLE tokens (
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    hash bytea PRIMARY KEY,
    scope text NOT NULL,
    expired_at timestamptz(0) NOT NULL,
    created_at timestamptz(0) NOT NULL
);