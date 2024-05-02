package apikeys

import (
	"fmt"
	"net/http"

	"github.com/fewsats/fewsatscli/client"
	"github.com/urfave/cli/v2"
)

var disableCommand = &cli.Command{
	Name:      "disable",
	Usage:     "Disable an API key.",
	ArgsUsage: "[api_key_id]",
	Action:    disableAPIKey,
}

func disableAPIKey(c *cli.Context) error {
	err := client.RequiresLogin()
	if err != nil {
		return cli.Exit("You need to log in to run this command.", 1)
	}

	if c.Args().Len() < 1 {
		return cli.Exit("API key ID is required", 1)
	}
	apiKeyID := c.Args().Get(0)

	client, err := client.NewHTTPClient()
	if err != nil {
		return cli.Exit("Failed to create HTTP client", 1)
	}

	endpoint := fmt.Sprintf("/v0/auth/apikeys/%s/disable", apiKeyID)
	resp, err := client.ExecuteRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		return cli.Exit("Failed to disable API key", 1)
	}
	defer resp.Body.Close()

	fmt.Println("API key disabled successfully")
	return nil
}
