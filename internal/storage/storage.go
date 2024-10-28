package storage

import (
	"bhsAssets/internal/config"
	"bhsAssets/internal/storage/models"
	"bhsAssets/internal/util"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
)

var (
	ErrNotFound = errors.New("not found")
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
	res, err := s.Db.Exec(stmt, balance, id)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}
	cnt, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}
	log.Println(cnt)
	if cnt != 1 {
		return fmt.Errorf("%s: User with id {%d} not found (cnt not match %d): %w,", fn, id, cnt, ErrNotFound)
	}
	return nil
}

func (s *Storage) CreateAsset(asset *models.Assets) (int64, error) {
	const fn = "storage.CreateAsset"

	var id int64
	stmt := `INSERT INTO assets(name, description, price, creator_id) VALUES($1, $2, $3, $4) RETURNING id`
	if err := s.Db.QueryRow(stmt, asset.Name, asset.Description, asset.Price, asset.CreatorId).Scan(&id); err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			if pgError.Code == pgerrcode.ForeignKeyViolation {
				return 0, fmt.Errorf("%s: User with id {%d} does not exist : %w", fn, asset.CreatorId, ErrNotFound)
			}
		}
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	return id, nil
}

func (s *Storage) GetAssetById(id int64) (*models.Assets, error) {
	const fn = "storage.GetAssetById"

	stmt := `SELECT * FROM assets WHERE id = $1`
	var asset models.Assets
	if err := s.Db.QueryRow(stmt, id).Scan(&asset.Id, &asset.Name, &asset.Description, &asset.Price, &asset.CreatorId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: Asset with id {%d} not found : %w", fn, id, ErrNotFound)
		}
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return &asset, nil
}

func (s *Storage) GetAllAssetsFiltered(filters map[string]string) ([]*models.Assets, error) {
	const fn = "storage.GetAllAssetsFiltered"
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar) // FOR POSTGRES !
	query := psql.Select("id, name, description, price, creator_id").From("assets")
	if name, ok := filters["name"]; ok {
		log.Printf("NAME IS %v \n", name)
		query = query.Where(squirrel.Expr("name LIKE ?", "%"+name+"%"))
	}

	if creatorId, ok := filters["creator_id"]; ok {
		query = query.Where(squirrel.Expr("creator_id = ?", creatorId))
	}

	if minPrice, ok := filters["min_price"]; ok {
		query = query.Where(squirrel.GtOrEq{"price": minPrice})
	}
	if maxPrice, ok := filters["max_price"]; ok {
		query = query.Where(squirrel.LtOrEq{"price": maxPrice})
	}
	sq, args, err := query.ToSql()

	if err != nil {
		return nil, fmt.Errorf("%s: error converting query to sql: %v", fn, err)
	}

	rows, err := s.Db.Query(sq, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: error quering data from db: %v", fn, err)
	}
	defer rows.Close()

	var assets []*models.Assets

	for rows.Next() {
		var asset models.Assets

		if err := rows.Scan(&asset.Id, &asset.Name, &asset.Description, &asset.Price, &asset.CreatorId); err != nil {
			return nil, fmt.Errorf("%s: error getting next row %v", fn, err)
		}
		assets = append(assets, &asset)
	}
	return assets, nil
}

func (s *Storage) BuyAsset(buyer *models.Users, asset *models.Assets, creator *models.Users) error {
	const fn = "storage.BuyAsset"
	tx, err := s.Db.Begin()
	if err != nil {
		return fmt.Errorf("%s: %v", fn, err)
	}
	defer tx.Rollback()

	var balance, price float64
	err = tx.QueryRow("SELECT balance FROM users WHERE id = $1 FOR UPDATE", buyer.Id).Scan(&balance)
	if err != nil {
		return fmt.Errorf("%s: failed to lock user row: %v", fn, err)
	}
	err = tx.QueryRow("SELECT price FROM assets WHERE id = $1", asset.Id).Scan(&price)
	if err != nil {
		return fmt.Errorf("%s: failed to lock asset row: %v", fn, err)
	}

	if balance < price {
		return fmt.Errorf("%s: insufficient balance", fn)
	}
	balance -= price
	_, err = tx.Exec("UPDATE users SET balance = $1 WHERE id = $2", balance, buyer.Id)
	if err != nil {
		return fmt.Errorf("%s: failed to update user balance: %v", fn, err)
	}

	_, err = tx.Exec("INSERT INTO assets_owners (owner_id, asset_id) VALUES ($1, $2)", buyer.Id, asset.Id)
	if err != nil {
		return fmt.Errorf("%s: failed to insert into assets_owners: %v", fn, err)
	}

	_, err = tx.Exec("UPDATE users SET balance = $1 WHERE id = $2", creator.Balance+price, creator.Id)
	if err != nil {
		return fmt.Errorf("%s: failed to update creator balance: %v", fn, err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: failed to commit transaction: %v", fn, err)
	}
	return nil
}

func (s *Storage) GetBoughtAssets(userId int64) ([]*models.Assets, error) {
	const fn = "storage.GetBoughtAssets"

	rows, err := s.Db.Query(`SELECT asset_id FROM assets_owners WHERE owner_id = $1`, userId)
	if err != nil {
		return nil, fmt.Errorf("%s: error converting query to sql: %v", fn, err)
	}
	defer rows.Close()

	var assets []*models.Assets

	for rows.Next() {
		var assetId int64
		var asset *models.Assets
		if err := rows.Scan(&assetId); err != nil {
			return nil, fmt.Errorf("%s: error quering asset id from db: %v", fn, err)
		}
		asset, assetGettingError := s.GetAssetById(assetId)
		if assetGettingError != nil {
			return nil, fmt.Errorf("%s: error getting asset with id %v: %v", fn, assetId, assetGettingError)
		}
		assets = append(assets, asset)
	}
	return assets, nil
}
