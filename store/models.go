package store

import "time"

type APIKey struct {
	ID        uint64     `db:"id"`
	Key       string     `db:"key"`
	Name      string     `db:"name"`
	HiddenKey string     `db:"hidden_key"`
	UserID    int64      `db:"user_id"`
	ExpiresAt *time.Time `db:"expires_at"`
	Enabled   bool       `db:"enabled"`
}
