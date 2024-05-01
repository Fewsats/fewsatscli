package store

import "time"

type APIKey struct {
	ID        uint64     `db:"id"`
	HiddenKey string     `db:"key"`
	ExpiresAt *time.Time `db:"expires_at"`
}
