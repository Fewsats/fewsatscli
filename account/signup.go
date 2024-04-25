package account

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"
	"syscall"

	"github.com/fewsats/fewsatscli/client"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

const (
	// signupPath is the path to the signup endpoint.
	signupPath = "/v0/auth/signup"
)

var signUpCommand = &cli.Command{
	Name:   "signup",
	Usage:  "Create a new account.",
	Flags:  []cli.Flag{},
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
	fmt.Print("Enter Email: ")
	var email string
	fmt.Scanln(&email)

	// Read password
	fmt.Print("Enter Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return cli.Exit("Failed to read password", 1)
	}
	password := strings.TrimSpace(string(bytePassword))

	// Read password confirmation
	fmt.Print("\nConfirm Password: ")
	bytePasswordConfirmation, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return cli.Exit("Failed to read password confirmation", 1)
	}
	passwordConfirmation := strings.TrimSpace(string(bytePasswordConfirmation))

	// Newline for the next prompt.
	fmt.Println()

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
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Debug(
				"Failed to read response body.",
				"error", err,
			)
		}

		slog.Debug(
			"Failed to create account.",
			"status_code", resp.StatusCode,
			"body", string(bodyBytes),
		)

		return cli.Exit("Failed to create account.", 1)
	}

	fmt.Println("Account created successfully.")

	return nil
}
