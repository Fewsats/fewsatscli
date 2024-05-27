package users

import (
	"github.com/urfave/cli/v2"
)

// BillingInformation represents a user's billing information.
type BillingInformation struct {
	// ID is the auto-incrementing ID of the billing information.
	ID uint64

	// UserID is the ID of the user the billing information belongs to.
	UserID uint64

	// FirstName is the first name of the user.
	FirstName string

	// LastName is the last name of the user.
	LastName string

	// AccountType is the type of account the user has.
	AccountType string

	// CompanyName is the name of the user's company.
	CompanyName string

	// Currency is the currency the user pays in.
	Currency string

	// Address is the street address of the billing address.
	Address string

	// City is the city of the billing address.
	City string

	// State is the state of the billing address.
	State string

	// Country is the country of the billing address.
	Country string

	// PostalCode is the postal code of the billing address.
	PostalCode string

	// VatNumber is the VAT number of the user.
	VatNumber string

	// TaxID is the tax ID of the user.
	TaxID string
}

func Command() *cli.Command {
	return &cli.Command{
		Name:  "users",
		Usage: "Interact with user information",
		Subcommands: []*cli.Command{
			getBillingCommand,
			updateBillingCommand,
		},
	}
}
