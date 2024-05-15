package credentials

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseL402Challenge(t *testing.T) {
	tests := []struct {
		name      string
		challenge string
		macaroon  string
		invoice   string
		expectErr string
	}{
		{
			name:      "Invalid challenge: empty header",
			challenge: "",
			macaroon:  "",
			invoice:   "",
			expectErr: "no L402 challenge/empty header found",
		},
		{
			name:      "Invalid challenge: missing macaroon",
			challenge: "L402 invoice=1234",
			macaroon:  "",
			invoice:   "",
			expectErr: "missing macaroon/invoice",
		},
		{
			name:      "Valid L402 challenge",
			challenge: "L402 macaroon=1234 invoice=1234",
			macaroon:  "1234",
			invoice:   "1234",
			expectErr: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc

			macaroon, invoice, err := parseL402Challenge(tc.challenge)
			if tc.expectErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.macaroon, macaroon)
			require.Equal(t, tc.invoice, invoice)
		})
	}
}
