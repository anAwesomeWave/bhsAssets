package storage

import (
	"bhsAssets/internal/config"
	"bhsAssets/internal/storage/models"
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
	Db *sql.DB // TODO: make this private. split and relocate migrator logic to this package
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

func (s *Storage) GetUser(login, password string) (*models.Users, error) {
	const fn = "storage.GetUser"
	stmt := `SELECT * FROM users WHERE login = $1`
	var user models.Users
	if err := s.Db.QueryRow(stmt, login).Scan(&user.Id, &user.Login, &user.PasswordHash, &user.Balance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: User {%s} with password {%s} not found : %w", fn, login, password, ErrNotFound)
		}
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	if !util.IsHashEqualPassword(user.PasswordHash, password) {
		return nil, fmt.Errorf("%s: User {%s} with password {%s} not found: passwords don't match : %w", fn, login, password, ErrNotFound)
	}

	return &user, nil
}

func (s *Storage) GetUserById(id int64) (*models.Users, error) {
	const fn = "storage.GetUser"
	stmt := `SELECT * FROM users WHERE id = $1`
	var user models.Users
	if err := s.Db.QueryRow(stmt, id).Scan(&user.Id, &user.Login, &user.PasswordHash, &user.Balance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: User with id {%d} not found : %w", fn, id, ErrNotFound)
		}
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return &user, nil
}

func (s *Storage) UpdateUserBalance(balance float64, id int64) error {
	const fn = "storage.GetUser"
	stmt := ` UPDATE users SET balance = $1 WHERE id = $2`
	var cnt int
	if err := s.Db.QueryRow(stmt, balance, id).Scan(&cnt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%s: User with id {%d} not found : %w", fn, id, ErrNotFound)
		}
		return fmt.Errorf("%s: %w", fn, err)
	}
	if cnt != 1 {
		return fmt.Errorf("%s: User with id {%d} not found : %w", fn, id, ErrNotFound)
	}
	return nil
}
