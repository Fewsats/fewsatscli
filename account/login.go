package account

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"syscall"
	"time"

	"github.com/fewsats/fewsatscli/apikeys"
	"github.com/fewsats/fewsatscli/client"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

const (
	// loginPath is the path to the login endpoint.
	loginPath = "/v0/auth/login"
)

// LoginRequest is the request body for the login endpoint.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var loginCommand = &cli.Command{
	Name:   "login",
	Usage:  "Log into your account.",
	Flags:  []cli.Flag{},
	Action: LoginCLI,
}

// LoginCLI handles the CLI login interaction.
func LoginCLI(c *cli.Context) error {
	fmt.Print("Enter email: ")
	var email string
	_, err := fmt.Scanln(&email)
	if err != nil {
		return cli.Exit("Failed to read email.", 1)
	}

	fmt.Print("Enter password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return cli.Exit("Failed to read password.", 1)
	}
	password := strings.TrimSpace(string(bytePassword))

	// Newline for the next prompt.
	fmt.Println()

	// Perform the login using the email and password
	_, err = Login(email, password)
	if err != nil {
		return cli.Exit("Login failed: "+err.Error(), 1)
	}

	fmt.Println("Login successful.")
	return nil
}

// Login to the fewsats API and return the session cookie
func Login(email, password string) (*http.Cookie, error) {
	method := http.MethodPost
	client, err := client.NewHTTPClient()
	if err != nil {
		slog.Debug(
			"Failed to create http client.",
			"error", err,
		)

		return nil, cli.Exit("Failed to create http client.", 1)
	}

	loginReqBody, err := json.Marshal(&LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		slog.Debug(
			"Failed to marshal login request body.",
			"error", err,
		)

		return nil, cli.Exit("Failed to marshal login request body.", 1)
	}

	resp, err := client.ExecuteRequest(method, loginPath, loginReqBody)
	if err != nil {
		slog.Debug(
			"Failed to execute login request.",
			"error", err,
		)

		return nil, cli.Exit("Failed to execute login request.", 1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Debug(
			"Login request failed.",
			"status_code", resp.StatusCode,
		)

		return nil, cli.Exit("Login request failed.", 1)
	}

	// Get the session cookie from the response
	var sessionCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "fewsats_session" {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		slog.Debug("Session cookie not found.")
		return nil, cli.Exit("Session cookie not found.", 1)
	}

	apiKey, expiresAt, err := apikeys.CreateAPIKey(24*7*4*time.Hour, "default", sessionCookie)
	if err != nil {
		slog.Debug("Failed to create API key on login.", "error", err)
		return nil, cli.Exit("Failed to create API key on login.", 1)
	}

	slog.Debug("API key created on login.")
	slog.Debug("API key:", "apiKey", apiKey)
	slog.Debug("Expires at:", "expiresAt", expiresAt)

	return sessionCookie, nil
}
