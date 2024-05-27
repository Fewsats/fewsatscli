package users

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"github.com/urfave/cli/v2"
	"github.com/fewsats/fewsatscli/client"
)

var updateBillingCommand = &cli.Command{
	Name:   "update-billing",
	Usage:  "Update billing information from a JSON file or JSON string",
	Action: updateBillingInformation,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "file",
			Usage: "Path to the JSON file containing billing information",
		},
		&cli.StringFlag{
			Name:  "json",
			Usage: "JSON string containing billing information",
		},
	},
}

func updateBillingInformation(c *cli.Context) error {
	var billingInfo BillingInformation
	if filePath := c.String("file"); filePath != "" {
		fileContent, err := ioutil.ReadFile(filePath)
		if err != nil {
			return cli.Exit(fmt.Sprintf("Failed to read file: %s", err), 1)
		}
		if err := json.Unmarshal(fileContent, &billingInfo); err != nil {
			return cli.Exit(fmt.Sprintf("Failed to unmarshal JSON from file: %s", err), 1)
		}
	} else if jsonString := c.String("json"); jsonString != "" {
		if err := json.Unmarshal([]byte(jsonString), &billingInfo); err != nil {
			return cli.Exit(fmt.Sprintf("Failed to unmarshal JSON string: %s", err), 1)
		}
	} else {
		return cli.Exit("Either --file or --json must be provided", 1)
	}

	reqBody, err := json.Marshal(billingInfo)
	if err != nil {
		return cli.Exit("Failed to marshal billing information", 1)
	}

	client, err := client.NewHTTPClient()
	if err != nil {
		return cli.Exit("Failed to create HTTP client", 1)
	}

	resp, err := client.ExecuteRequest(http.MethodPut, "/v0/users/billing", reqBody)
	if err != nil {
		return cli.Exit("Failed to update billing information", 1)
	}
	defer resp.Body.Close()

	fmt.Println("Billing information updated successfully.")
	return nil
}
