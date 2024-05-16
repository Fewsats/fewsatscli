package wallets

import (
	"errors"
	"fmt"

	"github.com/urfave/cli/v2"
)

//  connectWallet connects a new wallet with the given type.
func connectWallet(c *cli.Context) error {
	store, ok := c.App.Metadata["store"].(Store)
	if !ok {
		return errors.New("failed to get store from context")
	}

	walletType := c.String("type")

	switch walletType {
	case WalletTypeAlby, WalletTypeZBD:
		token := c.String("token")
		if token == "" {
			return fmt.Errorf("token argument is required for %s wallets",
				walletType)
		}

		_, err := connectTokenWallet(store, walletType, token)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported wallet type: %s", walletType)
	}

	return nil
}

// connectTokenWallet connects a new token based wallet with the given API key.
func connectTokenWallet(store Store, walletType, apiKey string) (uint64,
	error) {

	id, err := store.InsertWallet(walletType)
	if err != nil {
		return 0, fmt.Errorf("unable to insert wallet: %w", err)
	}

	err = store.InsertWalletToken(id, apiKey)
	if err != nil {
		return 0, fmt.Errorf("unable to insert wallet token: %w", err)
	}

	return id, nil
}
