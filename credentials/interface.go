package credentials

// Store is the interface that defines the methods that a credentials store
// should implement.
type Store interface {
	// GetL402Credentials retrieves the L402 credentials for a given service
	// from the database.
	GetL402Credentials(externalID string) (*L402Credentials, error)

	// InsertL402Credentials inserts the L402 credentials for a given service
	// into the database.
	InsertL402Credentials(challenge *L402Credentials) error
}
