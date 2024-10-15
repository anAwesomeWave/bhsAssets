-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS assets_owners (
      owner_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
      asset_id INTEGER REFERENCES assets(id) ON DELETE CASCADE,
      PRIMARY KEY (owner_id, asset_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE assets_owners;
-- +goose StatementEnd
