package store

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// newTestStore creates a new store for tests.
func newTestStore(t *testing.T) *Store {
	t.Helper()

	// Create a new store.
	db, err := sqlx.Connect("sqlite3", ":memory:")
	require.NoError(t, err)

	store := &Store{db: db}
	store.RunMigrations()

	return store
}
