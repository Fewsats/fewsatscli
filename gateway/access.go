package gateway

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/fewsats/fewsatscli/client"
	"github.com/fewsats/fewsatscli/config"
	"github.com/urfave/cli/v2"
)

const (
	accessGatewayPath = "/v0/gateway/access"
)

var accessCommand = &cli.Command{
	Name:      "access",
	Usage:     "Access a gateway endpoint.",
	ArgsUsage: "<gateway_id>",
	Action:    accessGateway,
}

func accessGateway(c *cli.Context) error {
	cfg, err := config.GetConfig()
	if err != nil {
		return cli.Exit("failed to get config", 1)
	}

	if c.Args().Len() < 1 {
		return cli.Exit("missing <gateway_id> argument", 1)
	}

	gatewayURL := c.Args().Get(0)
	// Check if the gatewayURL is a valid URL
	// If not, it means that the user used the gateway_id, prepend the domain to
	// create a valid URL
	_, err = url.ParseRequestURI(gatewayURL)
	if err != nil {
		gatewayURL = cfg.Domain + accessGatewayPath + "/" + gatewayURL
	}

	httpClient, err := client.NewHTTPClient()
	if err != nil {
		slog.Debug(
			"Failed to create http client.",
			"error", err,
		)
		return cli.Exit("failed to create http client", 1)
	}

	resp, err := httpClient.ExecuteL402Request(http.MethodGet, gatewayURL, nil)
	if err != nil {
		slog.Debug(
			"Failed to execute L402 request.",
			"error", err,
		)
		return cli.Exit("failed to execute request", 1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Debug(
			"Request failed",
			"status_code", resp.StatusCode,
		)
		return cli.Exit(fmt.Sprintf("failed to access gateway. Status code: %d", resp.StatusCode), 1)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Debug(
			"Failed to read response body",
			"error", err,
		)
		return cli.Exit("failed to read response body", 1)
	}

	// Print the response body
	fmt.Println(string(body))

	return nil
}
