package macaroons

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/fewsats/fewsatscli/client"
	"github.com/urfave/cli/v2"
)

type ValidateMacaroonRequest struct {
	Macaroon   string            `json:"macaroon"`
	Conditions map[string]string `json:"conditions"`
}

var validateCommand = &cli.Command{
	Name:   "validate",
	Usage:  "Validate a macaroon",
	Action: validateMacaroon,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "macaroon",
			Usage:    "Macaroon to validate",
			Required: true,
		},
		&cli.StringSliceFlag{
			Name:  "condition",
			Usage: "Conditions to validate against (can be used multiple times)",
		},
	},
}

func validateMacaroon(c *cli.Context) error {
	err := client.RequiresLogin()
	if err != nil {
		return cli.Exit("You need to log in to run this command.", 1)
	}

	macaroon := c.String("macaroon")
	conditions := c.StringSlice("condition")

	conditionMap := make(map[string]string)
	for _, condition := range conditions {
		key, value, found := strings.Cut(condition, "=")
		if !found {
			return cli.Exit(fmt.Sprintf("Invalid condition format: %s", condition), 1)
		}
		conditionMap[key] = value
	}

	req := ValidateMacaroonRequest{
		Macaroon:   macaroon,
		Conditions: conditionMap,
	}

	client, err := client.NewHTTPClient()
	if err != nil {
		slog.Debug("Failed to create HTTP client.", "error", err)
		return cli.Exit("Failed to create HTTP client.", 1)
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to marshal request: %v", err), 1)
	}

	resp, err := client.ExecuteRequest(http.MethodPost, "/v0/macaroon/validate", reqBody)
	if err != nil {
		slog.Debug("Failed to execute request.", "error", err)
		return cli.Exit("Failed to execute request.", 1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResponse struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			return cli.Exit(fmt.Sprintf("Failed to decode error response: %v", err), 1)
		}
		return cli.Exit(fmt.Sprintf("Failed to validate macaroon. %s", errorResponse.Error), 1)
	}

	var response struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return cli.Exit(fmt.Sprintf("Failed to decode response: %v", err), 1)
	}

	fmt.Println("Validation successful:", response.Message)
	return nil
}
