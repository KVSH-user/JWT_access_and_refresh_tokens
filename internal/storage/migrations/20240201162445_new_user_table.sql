-- +goose Up
CREATE TABLE IF NOT EXISTS users (
                                       id SERIAL PRIMARY KEY,
                                       refresh_token VARCHAR NOT NULL UNIQUE,
                                       guid VARCHAR NOT NULL,
                                       is_valid BOOLEAN NOT NULL DEFAULT TRUE,
                                       exp_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '30 days'

);

-- +goose Down
DROP TABLE users;