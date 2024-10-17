package storage

import (
	"bhsAssets/internal/config"
	"bhsAssets/internal/util"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	ErrNotFound = errors.New("not found ")
	ErrExists   = errors.New("alias already exists")
)

type Storage struct {
	Db *sql.DB
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

	return &Storage{Db: db}, nil
}

func (s *Storage) CreateUser(login string, password string) (int64, error) {
	const fn = "storage.CreateUser"

	pHash, err := util.GetHashPassword(password)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}
	var id int64
	stmt := `INSERT INTO users(login, password_hash) VALUES($1, $2) RETURNING id`

	if err := s.Db.QueryRow(stmt, login, pHash).Scan(&id); err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			if pgError.Code == pgerrcode.UniqueViolation {
				return 0, fmt.Errorf("%s: Login {%s} not unique : %w", fn, login, ErrExists)
			}
		}
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	return id, nil
}
