package wallets

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

//  connectWallet connects a new wallet with the given type.
func connectWallet(c *cli.Context) error {
	walletType := c.String("type")

	switch walletType {
	case WalletTypeAlby:
		token := c.String("token")
		if token == "" {
			return fmt.Errorf("token argument is required for Alby wallets")
		}

		_, err := ConnectAlbyWallet(token)
		if err != nil {
			return err
		}

	case WalletTypeZBD:
		token := c.String("token")
		if token == "" {
			return fmt.Errorf("token argument is required for Alby wallets")
		}

		_, err := ConnectAlbyWallet(token)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported wallet type: %s", walletType)
	}

	return nil
}
