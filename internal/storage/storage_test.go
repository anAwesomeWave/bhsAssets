package storage

import (
	"bhsAssets/internal/config"
	"bhsAssets/internal/util"
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"testing"
)

const storageEnvPath = "./../../config/.storage_env"
const storageInternalPath = "db:5432"

func TestStorageCreation(t *testing.T) {
	if err := godotenv.Load(storageEnvPath); err != nil {
		path, _ := os.Getwd()
		t.Fatalf(`Error with loading StorageEnv file: %v.
			!You should probably add proper config file with postgres params 
			(POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB)
			Test path: %s`, err, path)
	}
	t.Run("Empty config Storage Creation", func(t *testing.T) {
		emptyConfig := config.Storage{}
		if _, err := New(emptyConfig); err == nil {
			t.Fatal("Db created with empty config. Undefined behavior")
		}
	})

	t.Run("Default Storage Creation", func(t *testing.T) {
		storageCfg := *config.LoadStorage()
		storage, err := New(storageCfg)
		if err != nil {
			t.Fatalf("Cannot create storage with config %v: %v", storageCfg, err)
		}
		err = storage.Db.Ping()
		if err != nil {
			t.Fatalf("Cannot Ping created storage with config %v: %v", storageCfg, err)
		}
	})
}

func getStorageFixture() (*Storage, error) {
	const fn = "internal.storage.getStorageFixture"
	storageCfg := *config.LoadStorage()
	storage, err := New(storageCfg)
	if err != nil {
		return nil, fmt.Errorf("%s: Cannot create storage with config %v: %v", fn, storageCfg, err)
	}
	err = storage.Db.Ping()
	if err != nil {
		return nil, fmt.Errorf("%s: Cannot ping storage with config %v: %v", fn, storageCfg, err)
	}
	return storage, nil
}

func TestUserCreation(t *testing.T) {
	strg, err := getStorageFixture()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Create new User", func(t *testing.T) {
		login := "MyNewTestUser"
		password := "password"
		id, err := strg.CreateUser(login, password)
		if err != nil {
			t.Fatal(err)
		}
		var totalIds uint64
		if err := strg.Db.QueryRow(
			`SELECT COUNT(id) AS id_cnt FROM users WHERE id = $1`,
			id,
		).Scan(&totalIds); err != nil {
			t.Fatal(err)
		}
		if totalIds != 1 {
			t.Fatalf("Expected that there will be exactly 1 user with id %d, got %d", id, totalIds)
		}
	})
}

func TestGetUser(t *testing.T) {
	strg, err := getStorageFixture()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Get Created User", func(t *testing.T) {
		login := "test"
		password := "password"
		expBalance := 1000.0
		user, err := strg.GetUser(login, password)
		if err != nil {
			t.Fatal(err)
		}
		if user.Id <= 0 || user.Login != login || user.Balance != expBalance {
			t.Fatalf("Unexpected values from user's struct %v", *user)

		}
	})
	t.Run("Get User By ID", func(t *testing.T) {
		var id int64 = 1
		login := "test"
		password := "password"
		expBalance := 1000.0
		user, err := strg.GetUserById(id)
		if err != nil {
			t.Fatal(err)
		}
		if user.Id != id || user.Login != login || !util.IsHashEqualPassword(user.PasswordHash, password) ||
			user.Balance != expBalance {
			t.Fatalf("Unexpected values from user's struct %v", *user)

		}
	})
}
