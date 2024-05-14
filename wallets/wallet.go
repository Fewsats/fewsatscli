package wallets

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/fewsats/fewsatscli/store"
	"github.com/urfave/cli/v2"
)

const (
	WalletTypeAlby = "alby"
	WalletTypeZBD  = "zbd"
)

var (
	AllSupportedWallets = []string{
		WalletTypeAlby,
		WalletTypeZBD,
	}
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "wallet",
		Usage: "",
		Subcommands: []*cli.Command{
			ConnectWalletCommand,
		},
	}
}

var ConnectWalletCommand = &cli.Command{
	Name:  "connect",
	Usage: "Connect a new wallet",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name: "type",
			Usage: fmt.Sprintf("The type of the wallet: {%s}",
				strings.Join(AllSupportedWallets, ", ")),
			Required: true,
		},
		&cli.StringFlag{
			Name:  "token",
			Usage: "The token used to connect to the wallet",
		},
	},
	Action: connectWallet,
}

// GetDefaultWallet returns the default wallet in the database.
func GetDefaultWallet(logger *slog.Logger,
	store *store.Store) (PreimageProvider, error) {

	id, err := store.GetDefaultWallet()
	if err != nil {
		return nil, err
	}

	return GetWallet(logger, store, id)
}

// GetWallet returns a new wallet with the given type, if there is no wallet
// with the given type, it returns an error.
func GetWallet(logger *slog.Logger, store *store.Store,
	id uint64) (PreimageProvider, error) {

	wallet, err := store.GetWallet(id)
	if err != nil {
		return nil, err
	}

	var provider PreimageProvider
	switch wallet.Type {
	case "alby":
		token, err := store.GetWalletToken(id)
		if err != nil {
			return nil, err
		}

		provider = NewAlbyClient(token)

	case "zbd":
		token, err := store.GetWalletToken(id)
		if err != nil {
			return nil, err
		}

		provider = NewZBDClient(token)

	default:
		return nil, fmt.Errorf("unsupported wallet type: %s", wallet.Type)
	}

	return provider, nil
}

// DeleteWallet deletes the wallet from the database.
func DeleteWallet(id uint64, store *store.Store) error {
	wallet, err := store.GetWallet(id)
	if err != nil {
		return err
	}

	switch wallet.Type {
	case "alby":
		return DeleteAlbyWallet(id)

	case "zbd":
		return DeleteZBDWallet(id)

	default:
		return fmt.Errorf("delete wallet %s not implemented", wallet.Type)
	}

}
