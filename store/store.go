package store

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"database/sql"

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
		homeDir, _ := os.UserHomeDir()
		dbPath := filepath.Join(homeDir, ".fewsats", "fewsats.db")
		instance, _ = NewStore(dbPath)
		instance.InitSchema()
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
		expires_at DATETIME NOT NULL
	);`
	_, err := s.db.Exec(schema)
	return err
}

func (s *Store) InsertAPIKey(key string, expiresAt time.Time) (int64, error) {
	result, err := s.db.Exec("INSERT INTO api_keys (key, expires_at) VALUES (?, ?)", key, expiresAt)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
func (s *Store) GetAPIKey() (string, error) {
	var apiKey string
	err := s.db.Get(&apiKey, "SELECT key FROM api_keys WHERE expires_at > CURRENT_TIMESTAMP LIMIT 1")
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // Return an empty string and no error if no rows are found
		}
		return "", err
	}
	return apiKey, nil
}
