package credentials

import (
	"fmt"
	"net/http"
	"regexp"
	"time"
)

var (
	// ErrNoCredentialsFound is the error returned when no L402 credentials are
	// found in the database.
	ErrNoCredentialsFound = fmt.Errorf("no L402 credentials found")
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

// Precompiled regular expressions for performance
var (
	macaroonRegex = regexp.MustCompile(`macaroon="([^"]+)"`)
	invoiceRegex  = regexp.MustCompile(`invoice="([^"]+)"`)
)

// parseL402Challenge parses an L402 challenge and returns the macaroon and
// invoice.
func parseL402Challenge(challenge string) (string, string, error) {
	if challenge == "" {
		return "", "", fmt.Errorf("no L402 challenge/empty header found")
	}

	macaroonMatches := macaroonRegex.FindStringSubmatch(challenge)
	invoiceMatches := invoiceRegex.FindStringSubmatch(challenge)

	if macaroonMatches == nil || invoiceMatches == nil {
		return "", "", fmt.Errorf("missing macaroon/invoice in challenge: %s", challenge)
	}

	// Extracting the values from the regex matches
	macaroon := macaroonMatches[1]
	invoice := invoiceMatches[1]

	return macaroon, invoice, nil
}

// GetL402Credentials retrieves the L402 credentials from the database.
func GetL402Credentials(store Store, externalID string) (*L402Credentials,
	error) {

	creds, err := store.GetL402Credentials(externalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get L402 credentials from db: %w",
			err)
	}

	return creds, nil
}

func SaveL402Credentials(store Store, creds *L402Credentials) error {
	err := store.InsertL402Credentials(creds)
	if err != nil {
		return fmt.Errorf("failed to insert credentials to db: %w", err)
	}

	return nil
}
