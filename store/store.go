package store

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"sync"

	"github.com/fewsats/fewsatscli/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	"github.com/jmoiron/sqlx"

	_ "github.com/mattn/go-sqlite3"
)

var (
	instance *Store
	once     sync.Once
)

type Store struct {
	db *sqlx.DB
}

var (
	// ErrNoWalletFound is returned when no wallet is found in the database.
	ErrNoWalletFound = errors.New("no wallet found")
)

// GetStore returns the singleton instance of the store.
func GetStore() *Store {
	once.Do(func() {
		cfg, err := config.GetConfig()
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		instance, _ = NewStore(cfg.DBFilePath)
	})

	return instance
}

// NewStore creates a new store with the given database path.
func NewStore(dbPath string) (*Store, error) {
	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

// RunMigrations applies the database migrations to the latest version.
func (s *Store) RunMigrations() error {
	driver, err := sqlite3.WithInstance(s.db.DB, &sqlite3.Config{})
	if err != nil {
		return err
	}

	src, err := httpfs.New(http.FS(sqlSchemas), "migrations")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("httpfs", src, "sqlite3", driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			slog.Info("No migrations to run")
			return nil
		}

		return fmt.Errorf("unable to run migrations: %w", err)
	}

	return nil
}
