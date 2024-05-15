package credentials

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// L402Credentials is the struct that holds the L402 credentials of an L402
// challenge.
type L402Credentials struct {
	// ID is the unique identifier for the L402 credentials in the local
	// database.
	ID int64 `db:"id"`

	// ExternalID is the unique identifier for the related resource.
	ExternalID string `db:"external_id"`

	// Macaroon is the base64 encoded macaroon credentials.
	Macaroon string `db:"macaroon"`

	// Preimage is the preimage linked to this L402 challenge.
	Preimage string `db:"preimage"`

	// Invoice is the LN invoice linked to this L402 challenge.
	Invoice string `db:"invoice"`

	// CreatedAt is the time the L402 challenge stored in the database.
	CreatedAt time.Time `db:"created_at"`
}

// AuthenticationHeader returns the L402 header used to authenticated a request for the
// L402 credentials.
func (l *L402Credentials) AuthenticationHeader() (string, error) {
	// TODO(positiveblue): add credential validation.
	return fmt.Sprintf("L402 %s:%s", l.Macaroon, l.Preimage), nil
}

// ParseL402Challenge parses an L402 challenge from an HTTP request.
func ParseL402Challenge(externalID string,
	resp *http.Response) (*L402Credentials, error) {

	challenge := resp.Header.Get("WWW-Authenticate")

	macaroon, invoice, err := parseL402Challenge(challenge)
	if err != nil {
		return nil, fmt.Errorf("invalid L402 challenge header: %w", err)
	}

	return &L402Credentials{
		ExternalID: externalID,
		Macaroon:   macaroon,
		Invoice:    invoice,
	}, nil
}

// parseL402Challenge parses an L402 challenge and returns the macaroon and
// invoice.
func parseL402Challenge(challenge string) (string, string, error) {
	if challenge == "" {
		return "", "", fmt.Errorf("no L402 challenge/empty header found")
	}

	// Split the challenge into its parts.
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
		return "", "", fmt.Errorf("missing macaroon/invoice: %s", challenge)
	}

	return macaroon, invoice, nil
}
