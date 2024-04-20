package apikeys

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/fewsats/fewsatscli/client"
	"github.com/fewsats/fewsatscli/config"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/urfave/cli/v2"
)

const (
	// loginPath is the path to the login endpoint.
	loginPath = "/v0/auth/login"

	// apikeys is the path to the create/list api keys endpoint.
	apikeysPath = "/v0/auth/apikey"
)

var createCommand = &cli.Command{
	Name:  "new",
	Usage: "Create a new api key.",
	Flags: []cli.Flag{
		&cli.DurationFlag{
			Name:        "duration",
			Usage:       "The time duration for the api key to be valid in hours",
			Value:       24 * 7 * 4 * time.Hour,
			DefaultText: "28 days",
		},
	},

	Action: newApiKey,
}

// LoginRequest is the request body for the login endpoint.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CreateAPIKeyRequest is the request body for the create api key endpoint.
type CreateAPIKeyRequest struct {
	Duration time.Duration `json:"duration"`
}

// CreateAPIKeyResponse is the response body for the create api key endpoint.
type CreateAPIKeyResponse struct {
	APIKey    string     `json:"apikey"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// newApiKey creates a new api key.
func newApiKey(c *cli.Context) error {
	cfg, err := config.GetConfig()
	if err != nil {
		slog.Debug(
			"Failed to get config.",
			"error", err,
		)

		return cli.Exit("Failed to marshal JSON body.", 1)
	}

	duration := c.Duration("duration")

	req := &CreateAPIKeyRequest{Duration: duration}
	reqBody, err := json.Marshal(req)
	if err != nil {
		slog.Debug(
			"Failed to marshal JSON body.",
			"error", err,
		)

		return cli.Exit("Failed to marshal JSON body.", 1)
	}

	method := http.MethodPost
	client, err := client.NewHTTPClient()
	if err != nil {
		slog.Debug(
			"Failed to create http client.",
			"error", err,
		)

		return cli.Exit("Failed to create http client.", 1)
	}

	// If there is an API key in the config, use it to create a new API key.
	if cfg.APIKey != "" {
		resp, err := client.ExecuteRequest(method, apikeysPath, reqBody)
		if err != nil {
			slog.Debug(
				"Failed to execute request.",
				"error", err,
			)

			return cli.Exit("Failed to execute request.", 1)
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			slog.Debug(
				"Request failed",
				"status_code", resp.StatusCode,
			)

			return cli.Exit("Request failed.", 1)
		}

		var respData CreateAPIKeyResponse
		err = json.NewDecoder(resp.Body).Decode(&respData)
		if err != nil {
			slog.Debug(
				"Failed to decode response body.",
				"error", err,
			)

			return cli.Exit("Failed to decode response body.", 1)
		}

		fmt.Println("API key created.")
		fmt.Println("API key:", respData.APIKey)
		fmt.Println("Expires at:", respData.ExpiresAt)

		return nil
	}

	// If there is no API key in the config, we will get the user/password from
	// the user and create a new API key using a cookie session.
	fmt.Print("Login with your user account to create a new API key.\n")

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter email: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		slog.Debug(
			"Failed to read email.",
			"error", err,
		)

		return cli.Exit("Failed to read email.", 1)
	}

	fmt.Print("Enter password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		slog.Debug(
			"Failed to read password.",
			"error", err,
		)

		return cli.Exit("Failed to read password.", 1)
	}

	// Trim newline from the input
	email = strings.TrimSpace(email)
	password := strings.TrimSpace(string(bytePassword))

	// Newline for the next prompt.
	fmt.Println()

	loginReqBody, err := json.Marshal(&LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		slog.Debug(
			"Failed to marshal login request body.",
			"error", err,
		)

		return cli.Exit("Failed to marshal login request body.", 1)
	}

	resp, err := client.ExecuteRequest(method, loginPath, loginReqBody)
	if err != nil {
		slog.Debug(
			"Failed to execute login request.",
			"error", err,
		)

		return cli.Exit("Failed to execute login request.", 1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Debug(
			"Login request failed.",
			"status_code", resp.StatusCode,
		)

		return cli.Exit("Login request failed.", 1)
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
		return cli.Exit("Session cookie not found.", 1)
	}

	// Create a new request to create an API key
	httpReq, err := http.NewRequest(
		http.MethodPost, cfg.Domain+apikeysPath, bytes.NewBuffer(reqBody),
	)
	if err != nil {
		slog.Debug(
			"Failed to create request.",
			"error", err,
		)

		return cli.Exit("Failed to create request.", 1)
	}

	// Add the session cookie to the request
	httpReq.AddCookie(sessionCookie)

	// Send the request
	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		slog.Debug(
			"Failed to execute request.",
			"error", err,
		)

		return cli.Exit("Failed to execute request.", 1)
	}

	// Check the response
	if httpResp.StatusCode != http.StatusCreated {
		slog.Debug(
			"Request failed.",
			"status_code", resp.StatusCode,
		)

		return cli.Exit("Request failed.", 1)
	}

	// Parse the response body
	var respData CreateAPIKeyResponse
	err = json.NewDecoder(httpResp.Body).Decode(&respData)
	if err != nil {
		slog.Debug(
			"Failed to decode response body.",
			"error", err,
		)

		return cli.Exit("Failed to decode response body.", 1)
	}

	fmt.Println("API key created.")
	fmt.Println("API key:", respData.APIKey)
	fmt.Println("Expires at:", respData.ExpiresAt)

	return nil
}
