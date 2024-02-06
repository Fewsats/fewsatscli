package recipes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/urfave/cli/v2"
)

var ExecuteCodeRecipeCommand = &cli.Command{
	Name:  "execute",
	Usage: "Execute the code recipe with the defined parameters.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "recipe_id",
			Usage:    "The recipeID that will be executed.",
			Required: true,
		},
		&cli.StringSliceFlag{
			Name:     "data",
			Usage:    "The data to send in the POST request as key=value pairs.",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "alby_token",
			Usage:    "The access token for the Alby API.",
			Required: true,
		},
	},
	Action: executeCodeRecipe,
}

func extractInvoiceAndMacaroon(header string) (string, string, error) {
	// Remove the first 5 characters ("LSAT ")
	header = header[5:]
	// Split the header into parts
	parts := strings.Split(header, ", ")

	// Initialize empty invoice and macaroon
	var invoice, macaroon string

	// Loop over the parts and extract the invoice and macaroon
	for _, part := range parts {
		if strings.HasPrefix(part, "invoice=") {
			invoice = strings.TrimPrefix(part, "invoice=")
			invoice = strings.Trim(invoice, "\"")
		} else if strings.HasPrefix(part, "macaroon=") {
			macaroon = strings.TrimPrefix(part, "macaroon=")
			macaroon = strings.Trim(macaroon, "\"")
		}
	}

	// If either the invoice or macaroon is still empty, return an error
	if invoice == "" || macaroon == "" {
		return "", "", fmt.Errorf("could not extract invoice and macaroon from header")
	}

	return invoice, macaroon, nil
}

type PayInvoiceReq struct {
	Invoice string `json:"invoice"`
}

type PayInvoiceResp struct {
	PaymentPreimage string `json:"payment_preimage"`
}

func payInvoice(token, invoice string) (string, error) {
	// Create the invoice data
	data := PayInvoiceReq{
		Invoice: invoice,
	}

	// Marshal the data into JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	walletURL := fmt.Sprintf("%s/payments/bolt11", albyURL)

	// Create a new request
	req, err := http.NewRequest("POST", walletURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var respData PayInvoiceResp

	// Unmarshal the JSON response into the PayInvoiceResp object
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return "", err
	}

	return respData.PaymentPreimage, nil
}

type RecipeExecutionResponse struct {
	RecipeID      string                 `json:"recipe_id"`
	ExecutionTime float64                `json:"execution_time"`
	StatusCode    int                    `json:"status_code"`
	Result        map[string]interface{} `json:"result"`
}

// executeCodeRecipe executes a code recipe.
func executeCodeRecipe(c *cli.Context) error {
	recipeID := c.String("recipe_id")
	dataPairs := c.StringSlice("data")
	accessToken := c.String("alby_token")

	// Convert the data pairs to a map
	dataMap := make(map[string]string)
	for _, pair := range dataPairs {
		splitPair := strings.SplitN(pair, "=", 2)
		if len(splitPair) == 2 {
			dataMap[splitPair[0]] = splitPair[1]
		}
	}

	// Convert the map to JSON
	dataJSON, err := json.Marshal(dataMap)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/recipes/%s/execute", baseURL, recipeID)
	reqBody := bytes.NewBuffer(dataJSON)

	resp, err := http.Post(url, contentTypeJson, reqBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Make sure that this is an L402
	if resp.StatusCode != http.StatusPaymentRequired {
		return fmt.Errorf("unexpected status code: %d instead of %d", resp.StatusCode, http.StatusPaymentRequired)
	}

	// Extract the invoice and macaroon from the WWW-Authenticate header
	invoice, macaroon, err := extractInvoiceAndMacaroon(resp.Header.Get("WWW-Authenticate"))
	if err != nil {
		return fmt.Errorf("could not extract invoice and macaroon: %v", err)
	}

	fmt.Println("Paying L402...")
	// Pay the invoice
	preimage, err := payInvoice(accessToken, invoice)
	if err != nil {
		return fmt.Errorf("unable to pay invoice: %v", err)
	}

	// Create a new request
	finalReq, err := http.NewRequest("POST", url, reqBody)
	if err != nil {
		return fmt.Errorf("unable to create new request: %v", err)
	}

	finalReq.Header.Set("Authorization", fmt.Sprintf("LSAT %s:%s", macaroon, preimage))

	fmt.Println("Executing recipe...")
	// Send the request
	finalResp, err := http.DefaultClient.Do(finalReq)
	if err != nil {
		return err
	}
	defer finalResp.Body.Close()

	var recipeResp RecipeExecutionResponse
	err = json.NewDecoder(finalResp.Body).Decode(&recipeResp)
	if err != nil {
		return err
	}

	jsonBytes, err := json.MarshalIndent(recipeResp.Result, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println()
	fmt.Printf("Recipe ID: %s\n", recipeResp.RecipeID)
	fmt.Printf("Execution Time: %f\n", recipeResp.ExecutionTime)
	fmt.Printf("Status Code: %d\n", recipeResp.StatusCode)
	fmt.Printf("Result: %s\n", jsonBytes)
	fmt.Println()

	return nil
}
