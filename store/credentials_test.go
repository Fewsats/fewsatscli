package store

import (
	"testing"

	"github.com/fewsats/fewsatscli/credentials"
	"github.com/stretchr/testify/require"
)

func TestStoreCredentials(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)

	// The proper error is returned when the challenge is not found.
	_, err := store.GetL402Credentials("non-existent-uuid")
	require.Error(t, err)
	require.ErrorIs(t, err, credentials.ErrNoCredentialsFound)

	challenge := &credentials.L402Credentials{
		ExternalID: "externalID",
		Macaroon:   "Macaroon",
		Preimage:   "Preimage",
		Invoice:    "Invoice",
	}

	// Store the credentials.
	err = store.InsertL402Credentials(challenge)
	require.NoError(t, err)

	// Retrieve the credentials.
	dbChallenge, err := store.GetL402Credentials(challenge.ExternalID)
	require.NoError(t, err)

	// The db challenge should have an ID and a CreatedAt.
	require.NotZero(t, dbChallenge.ID)
	require.NotZero(t, dbChallenge.CreatedAt)

	// Populate the challenge with the ID and CreatedAt.
	challenge.ID = dbChallenge.ID
	challenge.CreatedAt = dbChallenge.CreatedAt

	// The challenges should match.
	require.Equal(t, challenge, dbChallenge)
}
