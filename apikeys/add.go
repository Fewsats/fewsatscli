package apikeys

import (
	"fmt"
	"time"

	"github.com/fewsats/fewsatscli/store"
	"github.com/urfave/cli/v2"
)

var addCommand = &cli.Command{
	Name:      "add",
	Usage:     "Add a new API key.",
	ArgsUsage: "[api_key]",
	Action:    addAPIKey,
}

func addAPIKey(c *cli.Context) error {
	if c.Args().Len() < 1 {
		return cli.Exit("API key is required", 1)
	}
	apiKey := c.Args().Get(0)

	store := store.GetStore()
	_, err := store.InsertAPIKey(apiKey, time.Now().Add(24*7*4*time.Hour), 0)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to add API key: %v", err), 1)
	}

	fmt.Println("API key added successfully.")
	return nil
}
