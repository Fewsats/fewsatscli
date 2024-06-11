package users

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/fewsats/fewsatscli/client"
	"github.com/urfave/cli/v2"
)

var getUserDetailsCommand = &cli.Command{
	Name:   "get-details",
	Usage:  "Get details of a user",
	Action: getUserDetails,
}

func getUserDetails(c *cli.Context) error {
	client, err := client.NewHTTPClient()
	if err != nil {
		slog.Debug("Failed to create HTTP client.", "error", err)
		return cli.Exit("Failed to create HTTP client", 1)
	}

	resp, err := client.ExecuteRequest(http.MethodGet, "/v0/users/details", nil)
	if err != nil {
		slog.Debug("Failed to get user details.", "error", err)
		return cli.Exit("Failed to get user details", 1)
	}
	defer resp.Body.Close()

	var userDetails User
	if err := json.NewDecoder(resp.Body).Decode(&userDetails); err != nil {
		slog.Debug("Failed to decode user details.", "error", err)
		return cli.Exit("Failed to decode user details", 1)
	}

	fmt.Printf("User Details: %+v\n", userDetails)
	return nil
}
