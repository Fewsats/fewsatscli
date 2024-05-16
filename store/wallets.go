package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/fewsats/fewsatscli/wallets"
)

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
			return 0, wallets.ErrNoWalletFound
		}

		return 0, err
	}

	return id, nil
}

// SetDefaultWallet sets the default wallet for the user.
func (s *Store) SetDefaultWallet(walletID uint64) error {
	stmt := `
		DELETE FROM default_wallet;
	`

	_, err := s.db.Exec(stmt)
	if err != nil {
		return fmt.Errorf("failed to delete default wallet: %w", err)
	}

	stmt = `
		INSERT INTO default_wallet (wallet_id)
		VALUES ($1);
	`

	_, err = s.db.Exec(stmt, walletID)
	if err != nil {
		return fmt.Errorf("failed to set default wallet: %w", err)
	}

	return nil
}

// InsertWallet inserts a new wallet in the database.
func (s *Store) InsertWallet(walletType string) (uint64, error) {
	stmt := `
		INSERT INTO wallets (wallet_type, created_at)
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
func (s *Store) GetWallet(id uint64) (*wallets.Wallet, error) {
	stmt := `
		SELECT id, wallet_type, created_at
		FROM wallets
		WHERE id = ?;
	`

	var wallet wallets.Wallet
	err := s.db.Get(&wallet, stmt, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, wallets.ErrNoWalletFound
		}

		return nil, err
	}

	return &wallet, nil
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
		WHERE wallet_id = ?;
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
