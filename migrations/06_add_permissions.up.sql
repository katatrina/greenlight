CREATE TABLE permissions (
    id bigserial PRIMARY KEY,
    code text NOT NULL
);

CREATE TABLE users_permissions (
    user_id bigint NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    permission_id bigint NOT NULL REFERENCES permissions (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, permission_id)
);

INSERT INTO permissions (code)
VALUES
    ('movies:read'),
    ('movies:write');