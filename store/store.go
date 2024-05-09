package store

import (
	"log"
	"sync"
	"time"

	"database/sql"

	"github.com/fewsats/fewsatscli/config"
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

func (s *Store) InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS api_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT NOT NULL,
		expires_at DATETIME NOT NULL,
		user_id INTEGER NOT NULL,
		enabled BOOLEAN NOT NULL DEFAULT 1
	);`
	_, err := s.db.Exec(schema)
	return err
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
