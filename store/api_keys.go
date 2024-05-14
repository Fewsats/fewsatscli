package store

import (
	"database/sql"
	"time"
)

// APIKey represents an API key in the database.
type APIKey struct {
	ID        uint64     `db:"id"`
	Key       string     `db:"key"`
	HiddenKey string     `db:"hidden_key"`
	UserID    int64      `db:"user_id"`
	ExpiresAt *time.Time `db:"expires_at"`
	Enabled   bool       `db:"enabled"`
}

// InsertAPIKey inserts a new API key into the database.
func (s *Store) InsertAPIKey(key string, expiresAt time.Time,
	userID int64) (int64, error) {

	stmt := `
		INSERT INTO api_keys (
			key, expires_at, user_id, enabled
		) VALUES (
			?, ?, ?, 1
		);
	`

	result, err := s.db.Exec(stmt, key, expiresAt, userID)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// GetAPIKey retrieves an "active" API key. An active API key is one that is
// enabled and has not expired.
//
// NOTE: If no active API keys are found, an empty string is returned.
func (s *Store) GetAPIKey() (string, error) {
	stmt := `
		SELECT key 
		FROM api_keys
		WHERE expires_at > CURRENT_TIMESTAMP AND enabled = 1
		LIMIT 1;
	`

	var apiKey string
	err := s.db.Get(&apiKey, stmt)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}

		return "", err
	}

	return apiKey, nil
}

// GetEnabledAPIKeys retrieves all enabled API keys that have not expired.
func (s *Store) GetEnabledAPIKeys() ([]APIKey, error) {
	stmt := `
		SELECT * 
		FROM api_keys
		WHERE expires_at > CURRENT_TIMESTAMP AND enabled = 1;
	`

	var apiKeys []APIKey
	err := s.db.Select(&apiKeys, stmt)
	return apiKeys, err
}

// DisableAPIKey sets the enabled field of an API key to false.
func (s *Store) DisableAPIKey(id uint64) error {
	stmt := `
		UPDATE api_keys 
		SET enabled = 0 
		WHERE id = ?;
	`
	_, err := s.db.Exec(stmt, id)
	return err
}
