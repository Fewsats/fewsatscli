package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/fewsats/fewsatscli/credentials"
)

// InsertL402Credentials inserts the L402 credentials into the database.
func (s *Store) InsertL402Credentials(
	challenge *credentials.L402Credentials) error {

	stmt := `
		INSERT INTO credentials (
			external_id, macaroon, preimage, invoice, created_at
		) VALUES (
			?, ?, ?, ?, ?
		);
	`

	challenge.CreatedAt = time.Now().UTC()

	_, err := s.db.Exec(
		stmt, challenge.ExternalID, challenge.Macaroon, challenge.Preimage,
		challenge.Invoice, challenge.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert L402 credentials: %w", err)
	}

	return nil
}

// GetL402Credentials retrieves the L402 credentials from the database.
func (s *Store) GetL402Credentials(
	externalID string) (*credentials.L402Credentials, error) {

	// TODO(positiveblue): make sure we are getting "valid" credentials:
	// not expired, etc...
	stmt := `
		SELECT *
		FROM credentials
		WHERE external_id = ?
		LIMIT 1;
	`

	var creds credentials.L402Credentials
	err := s.db.Get(&creds, stmt, externalID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, credentials.ErrNoCredentialsFound

	case err != nil:
		return nil, fmt.Errorf("failed to get L402 credentials for %s: %w",
			externalID, err)
	}

	if creds.Macaroon == "" || creds.Preimage == "" {
		return nil, fmt.Errorf("invalid L402 credentials for %s (empty "+
			"macaroon/preimage)", externalID)
	}

	return &creds, nil
}
