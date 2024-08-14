package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

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
			Name:     "amount",
			Usage:    "The amount in sats for the gateway",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "target-url",
			Usage:    "The URL to be proxied",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "duration",
			Usage:    "The duration for the gateway (e.g., '24h', '7d')",
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

	amount := c.Uint64("amount")
	targetURL := c.String("target-url")
	duration := c.String("duration")
	name := c.String("name")
	description := c.String("description")

	req := CreateGatewayRequest{
		Amount:      amount,
		TargetURL:   targetURL,
		Duration:    duration,
		Name:        name,
		Description: description,
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

	// Print raw response body for debugging
	bodyBytes, _ := io.ReadAll(resp.Body)
	fmt.Printf("Raw response: %s\n", string(bodyBytes))

	var gateway Gateway
	err = json.Unmarshal(bodyBytes, &gateway)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to decode response: %v. Raw response: %s", err, string(bodyBytes)), 1)
	}

	fmt.Println("Gateway created successfully:")
	fmt.Printf("Name: %s\n", gateway.Name)
	fmt.Printf("Description: %s\n", gateway.Description)
	fmt.Printf("Target URL: %s\n", gateway.TargetURL)
	// TODO(pol) this is not being returned as it's pricing's responsibility
	fmt.Printf("Amount: %d sats\n", gateway.Amount)
	fmt.Printf("Duration: %s\n", gateway.Duration)

	return nil
}

type CreateGatewayRequest struct {
	Amount      uint64 `json:"amount"`
	TargetURL   string `json:"target_url"`
	Duration    string `json:"duration"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Gateway struct {
	UserID      uint64    `json:"UserID"`
	ID          uint64    `json:"ID"`
	ExternalID  string    `json:"ExternalID"`
	TargetURL   string    `json:"TargetURL"`
	Status      string    `json:"Status"`
	Name        string    `json:"Name"`
	Description string    `json:"Description"`
	Amount      uint64    `json:"Amount"`
	Duration    string    `json:"Duration"`
	CreatedAt   time.Time `json:"CreatedAt"`
	UpdatedAt   time.Time `json:"UpdatedAt"`
}
