CREATE TABLE movies (
    id bigserial PRIMARY KEY,
    title text NOT NULL,
    runtime integer NOT NULL,
    genres text [] NOT NULL,
    year integer NOT NULL,
    version integer NOT NULL DEFAULT 1,
    created_at timestamptz(0) NOT NULL DEFAULT NOW()
);