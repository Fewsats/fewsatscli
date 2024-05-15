package wallets

// PreimageProvider is an interface for providing preimages for LN invoices.
type PreimageProvider interface {
	// GetPreimage returns the preimage for the given LN invoice.
	GetPreimage(invoice string) (string, error)
}

type Store interface {
	// GetDefaultWallet retrieves the default wallet ID.
	GetDefaultWallet() (uint64, error)
	// SetDefaultWallet sets the default wallet ID.
	SetDefaultWallet(walletID uint64) error

	// InsertWallet inserts a new wallet with the given type.
	InsertWallet(walletType string) (uint64, error)
	// GetWallet retrieves the wallet with the given ID.
	GetWallet(id uint64) (*Wallet, error)
	// GetWallets retrieves all wallets.
	DeleteWallet(id uint64) error

	// InsertWalletToken inserts a new token for the wallet with the given ID.
	InsertWalletToken(walletID uint64, token string) error
	// GetWalletToken retrieves the token for the wallet with the given ID.
	GetWalletToken(id uint64) (string, error)
	// DeleteWalletToken deletes the token for the wallet with the given ID.
	DeleteWalletToken(id uint64) error
}
