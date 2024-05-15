package wallets

import (
	"fmt"
)

// ConnectZBDWallet connects a new ZBD wallet with the given API key.
func ConnectZBDWallet(store Store, apiKey string) (uint64, error) {
	id, err := store.InsertWallet(WalletTypeZBD)
	if err != nil {
		return 0, fmt.Errorf("unable to insert wallet: %w", err)
	}

	err = store.InsertWalletToken(id, apiKey)
	if err != nil {
		return 0, fmt.Errorf("unable to insert wallet token: %w", err)
	}

	return id, nil
}

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
	return "", nil
}
