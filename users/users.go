package users

import (
	"github.com/urfave/cli/v2"
)

// User represents a platfomr user.
type User struct {
	// Email is the email address of the user. It is used as the unique
	// identifier for the user.
	Email string `json:"email"`

	// Username is the username of the user.
	Username string `json:"username"`

	// ProfileImageURL is the URL of the profile image of the user.
	ProfileImageURL string `json:"profile_image_url"`
}

// BillingInformation represents a user's billing information.
type BillingInformation struct {

	// FirstName is the first name of the user.
	FirstName string `json:"first_name"`

	// LastName is the last name of the user.
	LastName string `json:"last_name"`

	// AccountType is the type of account the user has.
	AccountType string `json:"account_type"`

	// CompanyName is the name of the user's company.
	CompanyName string `json:"company_name"`

	// Currency is the currency the user pays in.
	Currency string `json:"currency"`

	// Address is the street address of the billing address.
	Address string `json:"address"`

	// City is the city of the billing address.
	City string `json:"city"`

	// State is the state of the billing address.
	State string `json:"state"`

	// Country is the country of the billing address.
	Country string `json:"country"`

	// PostalCode is the postal code of the billing address.
	PostalCode string `json:"postal_code"`

	// VatNumber is the VAT number of the user.
	VatNumber string `json:"vat_number"`

	// TaxID is the tax ID of the user.
	TaxID string `json:"tax_id"`
}

func Command() *cli.Command {
	return &cli.Command{
		Name:  "users",
		Usage: "Interact with user information",
		Subcommands: []*cli.Command{
			getBillingCommand,
			updateBillingCommand,
			getUserDetailsCommand,
			updateUserDetailsCommand,
		},
	}
}
