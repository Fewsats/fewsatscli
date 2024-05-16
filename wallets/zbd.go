package wallets

import (
	"fmt"
)

// DeleteZBDWallet deletes the ZBD wallet with the given ID.
func DeleteZBDWallet(store Store, id uint64) error {
	err := store.DeleteWalletToken(id)
	if err != nil {
		return fmt.Errorf("unable to delete wallet token: %w", err)
	}

	err = store.DeleteWallet(id)

	return err
}

type ZBDClient struct {
	APIKey string
}

func NewZBDClient(apiKey string) *ZBDClient {
	return &ZBDClient{
		APIKey: apiKey,
	}
}

func (a *ZBDClient) GetPreimage(invoice string) (string, error) {
	return "", fmt.Errorf("not implemented")
}
