package storage

import (
	"bhsAssets/internal/config"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	ErrNotFound = errors.New("not found ")
	ErrExists   = errors.New("alias already exists")
)

type Storage struct {
	db *sql.DB
}

func New(storage config.Storage) (*Storage, error) {
	const fn = "storage.New"
	db, err := sql.Open(
		"pgx",
		fmt.Sprintf(
			"postgres://%s:%s@%s/%s?sslmode=disable",
			storage.User,
			storage.Password,
			storage.Path,
			storage.DbName),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: error pinging: %w", fn, err)
	}

	return &Storage{db: db}, nil
}
