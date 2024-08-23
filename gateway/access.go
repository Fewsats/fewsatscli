package gateway

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

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
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "method",
			Aliases: []string{"m"},
			Value:   "GET",
			Usage:   "HTTP method to use for the request (GET, POST, PUT, etc.)",
		},
		&cli.StringFlag{
			Name:    "body",
			Aliases: []string{"b"},
			Usage:   "Request body to send",
		},
		&cli.StringFlag{
			Name:  "content-type",
			Value: "application/json",
			Usage: "Content-Type header for the request",
		},
	},
	Action: accessGateway,
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

	method := c.String("method")
	body := c.String("body")
	contentType := c.String("content-type")

	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	resp, err := httpClient.ExecuteL402Request(method, gatewayURL, bodyReader, &contentType)
	if err != nil {
		slog.Debug(
			"Failed to execute L402 request.",
			"error", err,
			"method", method,
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
	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Debug(
			"Failed to read response body",
			"error", err,
		)
		return cli.Exit("failed to read response body", 1)
	}

	// Print the response body
	fmt.Println(string(bodyResp))

	return nil
}
