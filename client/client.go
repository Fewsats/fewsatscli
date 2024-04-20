package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/fewsats/fewsatscli/config"
	"github.com/lightningnetwork/lnd/zpay32"
)

// HttpClient is an HTTP client for interacting with the Fewsats API.
type HttpClient struct {
	client    *http.Client
	apiKey    string
	domain    string
	albyToken string
}

// NewHTTPClient creates a new HTTP client for interacting with the Fewsats API.
func NewHTTPClient() (*HttpClient, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to create http client: %w", err)
	}

	return &HttpClient{
		client:    &http.Client{},
		apiKey:    cfg.APIKey,
		domain:    cfg.Domain,
		albyToken: cfg.AlbyToken,
	}, nil
}

// ExecuteRequest executes an HTTP request with the given method, path, and body.
func (c *HttpClient) ExecuteRequest(method, path string,
	body []byte) (*http.Response, error) {

	url := fmt.Sprintf("%s%s", c.domain, path)
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	}

	if body != nil {
		req.Body = io.NopCloser(bytes.NewReader(body))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to execute request: %w", err)
	}

	return resp, nil
}

// ExecuteL402Request executes an HTTP request with the given method, path, and body.
// If the response status code is 402, it will show the user the price of the request
// and ask if they would like to proceed.
func (c *HttpClient) ExecuteL402Request(method, url string,
	body []byte) (*http.Response, error) {

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	}

	if body != nil {
		req.Body = io.NopCloser(bytes.NewReader(body))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to execute request: %w", err)
	}

	switch {
	case resp.StatusCode > 200 && resp.StatusCode < 300:
		return resp, nil

	case resp.StatusCode != http.StatusPaymentRequired:
		return resp, nil
	}

	fmt.Println("L402 Payment required received")
	fmt.Println("Auth header: ", resp.Header.Get("WWW-Authenticate"))
	reader := bufio.NewReader(os.Stdin)
	_, err = reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("unable to read user input: %w", err)
	}

	macaroon, invoice, err := ParseL402Challenge(resp)
	if err != nil {
		return nil, fmt.Errorf("unable to parse L402 challenge: %w", err)
	}

	invoicePrice, err := DecodePrice(invoice)
	if err != nil {
		return nil, fmt.Errorf("unable to decode invoice price: %w", err)
	}

	fmt.Printf("URL: %s\n", url)
	fmt.Printf("Lightning invoice price: %d sats\n", invoicePrice)
	fmt.Print("Do you want to continue? (Y/n): ")

	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("unable to read user input: %w", err)
	}

	input = strings.TrimSpace(input)
	if input != "Y" && input != "y" {
		return nil, fmt.Errorf("user chose not to continue")
	}

	preimage, err := PayInvoice(c.albyToken, invoice)
	if err != nil {
		return nil, fmt.Errorf("unable to pay invoice: %w", err)
	}

	fmt.Printf("Preimage: %s\n", preimage)

	req.Header.Set("Authorization", fmt.Sprintf("L402 %s:%s", macaroon, preimage))
	resp, err = c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to execute request: %w", err)
	}

	return resp, nil
}

// ParseL402Challenge parses an L402 challenge from an HTTP response.
func ParseL402Challenge(resp *http.Response) (string, string, error) {
	challenge := resp.Header.Get("WWW-Authenticate")
	if challenge == "" {
		return "", "", fmt.Errorf("no L402 challenge found")
	}

	parts := strings.Split(challenge, " ")

	var macaroon, invoice string
	for _, part := range parts {
		if strings.HasPrefix(part, "macaroon=") {
			macaroon = strings.TrimPrefix(part, "macaroon=")
		} else if strings.HasPrefix(part, "invoice=") {
			invoice = strings.TrimPrefix(part, "invoice=")
		}
	}

	if macaroon == "" || invoice == "" {
		return "", "", fmt.Errorf("macaroon or invoice not found in challenge")
	}

	return macaroon, invoice, nil
}

// PaymentResponse represents a payment response.
type PaymentResponse struct {
	Amount          int    `json:"amount"`
	Description     string `json:"description"`
	Destination     string `json:"destination"`
	Fee             int    `json:"fee"`
	PaymentHash     string `json:"payment_hash"`
	PaymentPreimage string `json:"payment_preimage"`
	PaymentRequest  string `json:"payment_request"`
}

// PayInvoice pays a lightning invoice.
func PayInvoice(accessToken, invoice string) (string, error) {
	url := "https://api.getalby.com/payments/bolt11"

	// Create the request body
	body := map[string]interface{}{
		"invoice": invoice,
	}
	reqBodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("unable to encode request body: %w", err)
	}

	// Create the request
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return "", fmt.Errorf("unable to create request: %w", err)
	}

	// Set the Authorization header
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("unable to send request: %w", err)
	}
	// Parse the response body
	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("unable to read response body: %w", err)
	}

	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var paymentResponse PaymentResponse
	err = json.Unmarshal(respBodyBytes, &paymentResponse)
	if err != nil {
		return "", fmt.Errorf("unable to parse response body: %w", err)
	}

	return paymentResponse.PaymentPreimage, nil
}

// DecodePrice decodes a price from a ln payment request.
func DecodePrice(invoice string) (uint64, error) {
	if len(invoice) < 2 {
		return 0, errors.New("bolt11 too short")
	}

	firstNumber := strings.IndexAny(invoice, "1234567890")
	if firstNumber < 2 {
		return 0, errors.New("invalid bolt11 invoice")
	}

	chainPrefix := strings.ToLower(invoice[2:firstNumber])
	chain := &chaincfg.Params{
		Bech32HRPSegwit: chainPrefix,
	}

	inv, err := zpay32.Decode(invoice, chain)
	if err != nil {
		return 0, fmt.Errorf("zpay32 decoding failed: %w", err)
	}

	var msat int64
	if inv.MilliSat != nil {
		msat = int64(*inv.MilliSat)
	}

	return uint64(msat / 1000), nil
}
