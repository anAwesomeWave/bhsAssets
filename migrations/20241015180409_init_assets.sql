-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS assets (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    balance NUMERIC DEFAULT 0.0,
    creator_id INTEGER REFERENCES users(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE assets;
-- +goose StatementEnd
