package account

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/mail"

	"github.com/fewsats/fewsatscli/client"
	"github.com/urfave/cli/v2"
)

const (
	// signupPath is the path to the signup endpoint.
	signupPath = "/v0/auth/signup"
)

var signUpCommand = &cli.Command{
	Name:  "signup",
	Usage: "Create a new account.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "email",
			Usage:    "The email address of the account.",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "password",
			Usage:    "The password of the account.",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "password-confirmation",
			Usage:    "The password confirmation, must match the password.",
			Required: true,
		},
	},
	Action: signup,
}

// SignupRequest is the request body for the signup endpoint.
type SignupRequest struct {
	Email                string `json:"email"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password2"`
}

// signup creates a new account.
func signup(c *cli.Context) error {
	email := c.String("email")
	password := c.String("password")
	passwordConfirmation := c.String("password-confirmation")

	// Check if the email address is valid.
	if _, err := mail.ParseAddress(email); err != nil {
		return cli.Exit("The email address is invalid.", 1)
	}

	// Check if the password and password confirmation match.
	if password != passwordConfirmation {
		return cli.Exit(
			"The password and password confirmation do not match.", 1,
		)
	}

	req := SignupRequest{
		Email:                email,
		Password:             password,
		PasswordConfirmation: passwordConfirmation,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		slog.Debug(
			"Failed to marshal JSON body.",
			"error", err,
		)

		return cli.Exit("Failed to marshal JSON body.", 1)
	}

	client, err := client.NewHTTPClient()
	if err != nil {
		slog.Debug(
			"Failed to create HTTP client.",
			"error", err,
		)

		return cli.Exit("Failed to create HTTP client.", 1)
	}

	resp, err := client.ExecuteRequest(http.MethodPost, signupPath, reqBody)
	if err != nil {
		return cli.Exit(err.Error(), 1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		slog.Debug(
			"Failed to create account.",
			"status_code", resp.StatusCode,
		)

		return cli.Exit("Failed to create account.", 1)
	}

	fmt.Println("Account created successfully.")

	return nil
}
