package apikeys

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/fewsats/fewsatscli/client"
	"github.com/fewsats/fewsatscli/store"

	"github.com/urfave/cli/v2"
)

const (
	// apikeysPath is the path to the create/list api keys endpoint.
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

// CreateAPIKeyRequest is the request body for the create api key endpoint.
type CreateAPIKeyRequest struct {
	Duration time.Duration `json:"duration"`
}

// CreateAPIKeyResponse is the response body for the create api key endpoint.
type CreateAPIKeyResponse struct {
	APIKey    string     `json:"apikey"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// CreateAPIKey is an exported function to create an API key.
func CreateAPIKey(duration time.Duration, sessionCookie *http.Cookie) (string, *time.Time, error) {
	req := &CreateAPIKeyRequest{Duration: duration}
	reqBody, err := json.Marshal(req)
	if err != nil {
		slog.Debug("Failed to marshal JSON body.", "error", err)
		return "", nil, err
	}

	client, err := client.NewHTTPClient()
	if err != nil {
		slog.Debug("Failed to create http client.", "error", err)
		return "", nil, err
	}

	if sessionCookie != nil {
		client.SetSessionCookie(sessionCookie)
	}

	resp, err := client.ExecuteRequest(http.MethodPost, apikeysPath, reqBody)
	if err != nil {
		slog.Debug("Failed to execute request.", "error", err)
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	var respData CreateAPIKeyResponse
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		slog.Debug("Failed to decode response body.", "error", err)
		return "", nil, err
	}

	// Store the API key in the database
	store := store.GetStore()
	_, err = store.InsertAPIKey(respData.APIKey, *respData.ExpiresAt)
	if err != nil {
		slog.Debug("Failed to insert API key into database.", "error", err)
		return "", nil, err
	}

	return respData.APIKey, respData.ExpiresAt, nil
}

// newApiKey creates a new api key.
func newApiKey(c *cli.Context) error {
	duration := c.Duration("duration")
	apiKey, expiresAt, err := CreateAPIKey(duration, nil)
	if err != nil {
		return cli.Exit(err.Error(), 1)
	}

	fmt.Println("API key created.")
	fmt.Println("API key:", apiKey)
	fmt.Println("Expires at:", expiresAt)

	return nil
}
