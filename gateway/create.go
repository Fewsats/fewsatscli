package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/fewsats/fewsatscli/client"
	"github.com/urfave/cli/v2"
)

const (
	createGatewayPath = "/v0/gateway"
)

var createCommand = &cli.Command{
	Name:  "create",
	Usage: "Create a new gateway.",
	Flags: []cli.Flag{
		&cli.Uint64Flag{
			Name:     "price",
			Usage:    "The price in cents for the gateway",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "target-url",
			Usage:    "The URL to be proxied",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "name",
			Usage:    "The name of the gateway",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "description",
			Usage:    "The description of the gateway",
			Required: true,
		},
	},
	Action: createGateway,
}

func createGateway(c *cli.Context) error {
	err := client.RequiresLogin()
	if err != nil {
		return cli.Exit("You need to log in to run this command.", 1)
	}

	priceInCents := c.Uint64("price")
	targetURL := c.String("target-url")
	name := c.String("name")
	description := c.String("description")

	req := CreateGatewayRequest{
		PriceInCents: priceInCents,
		TargetURL:    targetURL,
		Name:         name,
		Description:  description,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to marshal request: %v", err), 1)
	}

	client, err := client.NewHTTPClient()
	if err != nil {
		slog.Debug("Failed to create HTTP client.", "error", err)
		return cli.Exit("Failed to create HTTP client.", 1)
	}

	resp, err := client.ExecuteRequest(http.MethodPost, createGatewayPath, jsonData)
	if err != nil {
		slog.Debug("Failed to execute request.", "error", err)
		return cli.Exit("Failed to execute request.", 1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return cli.Exit(fmt.Sprintf("Failed to create gateway. Status code: %d, Body: %s", resp.StatusCode, string(bodyBytes)), 1)
	}

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to read response body: %v", err), 1)
	}

	// Unmarshal the response into a Gateway struct
	var gateway Gateway
	err = json.Unmarshal(bodyBytes, &gateway)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to decode response: %v. Raw response: %s", err, string(bodyBytes)), 1)
	}

	// Create a response struct to match the format of other commands
	response := struct {
		Gateway Gateway `json:"gateway"`
	}{
		Gateway: gateway,
	}

	// Marshal the response struct to JSON
	jsonOutput, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return cli.Exit("Failed to marshal JSON.", 1)
	}

	fmt.Println(string(jsonOutput))

	return nil
}

type CreateGatewayRequest struct {
	PriceInCents uint64 `json:"price_in_cents"`
	TargetURL    string `json:"target_url"`
	Duration     string `json:"duration"`
	Name         string `json:"name"`
	Description  string `json:"description"`
}
