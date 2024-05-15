package wallets

import (
	"fmt"
	"strings"
	"time"

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

	ErrNoWalletFound = fmt.Errorf("no wallet found")
)

// Wallet represents a connected wallet able to provide preimages for LN
// invoices.
type Wallet struct {
	ID        uint64    `db:"id"`
	Type      string    `db:"wallet_type"`
	CreatedAt time.Time `db:"created_at"`
}

// WalletToken represents a wallet that can be accessed by a token, like an API
// driven wallet like Alby or ZBD.
type WalletToken struct {
	ID       uint64 `db:"id"`
	WalletID uint64 `db:"wallet_id"`
	Token    string `db:"token"`
}

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
func GetDefaultWallet(store Store) (PreimageProvider,
	error) {

	id, err := store.GetDefaultWallet()
	if err != nil {
		return nil, err
	}

	return GetWallet(store, id)
}

// GetWallet returns a new wallet with the given type, if there is no wallet
// with the given type, it returns an error.
func GetWallet(store Store, id uint64) (PreimageProvider, error) {

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
func DeleteWallet(store Store, id uint64) error {
	wallet, err := store.GetWallet(id)
	if err != nil {
		return err
	}

	switch wallet.Type {
	case "alby":
		return DeleteAlbyWallet(store, id)

	case "zbd":
		return DeleteZBDWallet(store, id)

	default:
		return fmt.Errorf("delete wallet %s not implemented", wallet.Type)
	}
}
