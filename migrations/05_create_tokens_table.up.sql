CREATE TABLE tokens (
    hash bytea PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    expiry timestamptz(0) NOT NULL,
    scope text NOT NULL
);