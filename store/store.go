package store

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"database/sql"

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
			slog.Debug("No migrations to run")
			return nil
		}

		return fmt.Errorf("unable to run migrations: %w", err)
	}

	return nil
}

func (s *Store) InsertAPIKey(key string, expiresAt time.Time, userID int64) (int64, error) {
	result, err := s.db.Exec("INSERT INTO api_keys (key, expires_at, user_id, enabled) VALUES (?, ?, ?, 1)", key, expiresAt, userID)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *Store) GetAPIKey() (string, error) {
	var apiKey string
	err := s.db.Get(&apiKey, "SELECT key FROM api_keys WHERE expires_at > CURRENT_TIMESTAMP AND enabled = 1 LIMIT 1")
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // Return an empty string and no error if no rows are found
		}
		return "", err
	}
	return apiKey, nil
}

// GetEnabledAPIKeys retrieves all enabled API keys that have not expired.
func (s *Store) GetEnabledAPIKeys() ([]APIKey, error) {
	var apiKeys []APIKey
	err := s.db.Select(&apiKeys, "SELECT * FROM api_keys WHERE expires_at > CURRENT_TIMESTAMP AND enabled = 1")
	return apiKeys, err
}

// DisableAPIKey sets the enabled field of an API key to false.
func (s *Store) DisableAPIKey(id uint64) error {
	_, err := s.db.Exec("UPDATE api_keys SET enabled = 0 WHERE id = ?", id)
	return err
}
