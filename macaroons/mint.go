package macaroons

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/fewsats/fewsatscli/client"
	"github.com/urfave/cli/v2"
)

type MintMacaroonRequest struct {
	Location string            `json:"location"`
	Caveats  map[string]string `json:"caveats"`
}

var mintCommand = &cli.Command{
	Name:   "mint",
	Usage:  "Mint a new macaroon",
	Action: mintMacaroon,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "location",
			Usage:    "Location for the macaroon",
			Required: true,
		},
		&cli.StringSliceFlag{
			Name:  "caveat",
			Usage: "Caveats for the macaroon (can be used multiple times)",
		},
	},
}

type MacaroonResponse struct {
	Macaroon string `json:"macaroon"`
}

func mintMacaroon(c *cli.Context) error {
	err := client.RequiresLogin()
	if err != nil {
		return cli.Exit("You need to log in to run this command.", 1)
	}

	location := c.String("location")
	caveats := c.StringSlice("caveat")

	caveatMap := make(map[string]string)
	for _, caveat := range caveats {
		key, value, found := strings.Cut(caveat, "=")
		if !found {
			return cli.Exit(fmt.Sprintf("Invalid caveat format: %s", caveat), 1)
		}
		caveatMap[key] = value
	}

	req := MintMacaroonRequest{
		Location: location,
		Caveats:  caveatMap,
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

	resp, err := client.ExecuteRequest(http.MethodPost, "/v0/macaroon/mint", reqBody)
	if err != nil {
		slog.Debug("Failed to execute request.", "error", err)
		return cli.Exit("Failed to execute request.", 1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return cli.Exit(fmt.Sprintf("Failed to mint macaroon. Status code: %d", resp.StatusCode), 1)
	}

	// Read the entire response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to read response body: %v", err), 1)
	}
	var response MacaroonResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return cli.Exit(fmt.Sprintf("Failed to decode response: %v", err), 1)
	}

	jsonOutput, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return cli.Exit("Failed to marshal JSON.", 1)
	}

	fmt.Println(string(jsonOutput))

	return nil
}
