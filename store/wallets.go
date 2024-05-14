package store

import (
	"database/sql"
	"errors"
	"time"
)

// Wallet represents a connected wallet able to provide preimages for LN
// invoices.
type Wallet struct {
	ID        uint64    `db:"id"`
	Type      string    `db:"type"`
	CreatedAt time.Time `db:"created_at"`
}

// WalletToken represents a wallet that can be accessed by a token, like an API
// driven wallet like Alby or ZBD.
type WalletToken struct {
	WalletID uint64 `db:"wallet_id"`
	Token    string `db:"token"`
}

// GetDefaultWallet returns the default wallet in the database.
func (s *Store) GetDefaultWallet() (uint64, error) {
	stmt := `
		SELECT wallet_id
		FROM default_wallet;
	`

	var id uint64
	err := s.db.Get(&id, stmt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrNoWalletFound
		}

		return 0, err
	}

	return id, nil
}

// SetDefaultWallet sets the default wallet for the user.
func (s *Store) SetDefaultWallet(walletID uint64) error {
	stmt := `
		INSERT INTO default_wallet (wallet_id)
		VALUES ($1)
		ON CONFLICT (wallet_id) DO UPDATE
		SET wallet_id = EXCLUDED.wallet_id;
	`

	_, err := s.db.Exec(stmt, walletID)
	if err != nil {
		return err
	}

	return nil
}

// InsertWallet inserts a new wallet in the database.
func (s *Store) InsertWallet(walletType string) (uint64, error) {
	stmt := `
		INSERT INTO wallets (type, created_at)
		VALUES ($1, $2)
		RETURNING id;
	`

	var id uint64
	err := s.db.Get(&id, stmt, walletType, time.Now().UTC())
	if err != nil {
		return 0, err
	}

	err = s.SetDefaultWallet(id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetWallet returns the wallet stored in the database.
func (s *Store) GetWallet(id uint64) (*Wallet, error) {
	stmt := `
		SELECT id, type, created_at
		FROM wallets
		WHERE id = $1;
	`

	var wallet *Wallet
	err := s.db.Get(&wallet, stmt, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoWalletFound
		}
	}

	return wallet, nil
}

// DeleteWallet deletes the wallet with the given ID from the database.
func (s *Store) DeleteWallet(id uint64) error {
	stmt := `
		DELETE
		FROM wallets
		WHERE id = $1;
	`

	_, err := s.db.Exec(stmt, id)
	if err != nil {
		return err
	}

	return nil
}

// InsertWalletToken inserts a new wallet token in the database.
func (s *Store) InsertWalletToken(walletID uint64, token string) error {
	stmt := `
		INSERT INTO token_based_preimage_provider (wallet_id, token)
		VALUES ($1, $2);
	`

	_, err := s.db.Exec(stmt, walletID, token)
	if err != nil {
		return err
	}

	return nil
}

// GetWalletToken returns the token for the wallet stored in the database.
func (s *Store) GetWalletToken(id uint64) (string, error) {
	stmt := `
		SELECT token
		FROM token_based_preimage_provider
		WHERE wallet_id = $1;
	`

	var token string
	err := s.db.Get(&token, stmt, id)
	if err != nil {
		return "", err
	}

	return token, nil
}

// DeleteWalletToken deletes the wallet token with the given ID from the database.
func (s *Store) DeleteWalletToken(id uint64) error {
	stmt := `
		DELETE
		FROM token_based_preimage_provider
		WHERE wallet_id = $1;
	`

	_, err := s.db.Exec(stmt, id)
	if err != nil {
		return err
	}

	return nil
}
