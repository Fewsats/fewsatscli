package store

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// L402Credentials is the struct that holds the L402 credentials of an L402
// challenge.
type L402Credentials struct {
	// ExternalID is the unique identifier for the related resource.
	ExternalID string

	// Macaroon is the base64 encoded macaroon credentials.
	Macaroon string `db:"macaroon"`

	// Preimage is the preimage linked to this L402 challenge.
	Preimage string `db:"preimage"`

	// Invoice is the LN invoice linked to this L402 challenge.
	Invoice string `db:"invoice"`

	// CreatedAt is the time the L402 challenge stored in the database.
	CreatedAt time.Time `db:"created_at"`
}

// L402Header returns the L402 header used to authenticated a request for the
// L402 credentials.
func (l *L402Credentials) L402Header() string {
	return fmt.Sprintf("L402 %s:%s", l.Macaroon, l.Preimage)
}

// ParseL402Challenge parses an L402 challenge from an HTTP response.
func ParseL402Challenge(externalID string,
	resp *http.Response) (*L402Credentials, error) {

	challenge := resp.Header.Get("WWW-Authenticate")
	if challenge == "" {
		return nil, fmt.Errorf("no L402 challenge found")
	}

	parts := strings.Split(challenge, " ")

	var macaroon, invoice string
	for _, part := range parts {
		if strings.HasPrefix(part, "macaroon=") {
			macaroon = strings.TrimPrefix(part, "macaroon=")
		} else if strings.HasPrefix(part, "invoice=") {
			invoice = strings.TrimPrefix(part, "invoice=")
		}
	}

	if macaroon == "" || invoice == "" {
		return nil, fmt.Errorf("macaroon or invoice not found in challenge: %s",
			challenge)
	}

	return &L402Credentials{
		ExternalID: externalID,
		Macaroon:   macaroon,
		Invoice:    invoice,
	}, nil
}

// InsertL402Credentials
func (s *Store) InsertL402Credentials(challenge *L402Credentials) error {
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
func (s *Store) GetL402Credentials(externalID string) (*L402Credentials,
	error) {

	// TODO(positiveblue): make sure we are getting "valid" credentials:
	// not expired, etc...
	stmt := `
		SELECT macaroon, preimage, invoice, created_at
		FROM credentials
		WHERE external_id = ?
		LIMIT 1;
	`

	var credentials L402Credentials
	err := s.db.Get(&credentials, stmt, externalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get L402 credentials for %s: %w",
			externalID, err)
	}

	if credentials.Macaroon == "" || credentials.Preimage == "" {
		return nil, fmt.Errorf("invalid L402 credentials for %s (empty "+
			"macaroon/preimage)", externalID)
	}

	return &credentials, nil
}
