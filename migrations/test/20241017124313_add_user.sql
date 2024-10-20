-- +goose Up
-- +goose StatementBegin
INSERT INTO users(login, password_hash, balance)
    VALUES ('test', '$2a$10$XxSQL4UOS5VfrveegUhaouZT4z1dbcVWj0l/AZPmWieyxpBAVocLu', 1000.0);  -- hash of "password" get by calling
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DELETE FROM users
    WHERE login = "test";
-- +goose StatementEnd
