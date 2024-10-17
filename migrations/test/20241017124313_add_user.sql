-- +goose Up
-- +goose StatementBegin
INSERT INTO users(login, password_hash)
    VALUES ('test', '\x243261243130246A7234576130657A30695A5973654750627178317975727852586B6C43396F6E30426F52776C4A316B692F6843724A4E486E484E4F');  -- hash of "password" get by calling `hex.EncodeToString(pHash)`
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DELETE FROM users
    WHERE login = "test";
-- +goose StatementEnd
