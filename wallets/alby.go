package wallets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/fewsats/fewsatscli/store"
)

const (
	albyURL = "https://api.getalby.com"
)

// ConnectAlbyWallet connects a new Alby wallet with the given API key.
func ConnectAlbyWallet(apiKey string) (uint64, error) {
	store := store.GetStore()
	id, err := store.InsertWallet(WalletTypeAlby)
	if err != nil {
		return 0, fmt.Errorf("unable to insert wallet: %w", err)
	}

	err = store.InsertWalletToken(id, apiKey)
	if err != nil {
		return 0, fmt.Errorf("unable to insert wallet token: %w", err)
	}

	return id, nil
}

// DeleteAlbyWallet deletes the Alby wallet with the given ID.
func DeleteAlbyWallet(id uint64) error {
	store := store.GetStore()

	err := store.DeleteWalletToken(id)
	if err != nil {
		return fmt.Errorf("unable to delete wallet token: %w", err)
	}

	err = store.DeleteWallet(id)

	return err
}

// AlbyClient is a client for the Alby wallet API
type AlbyClient struct {
	// APIKey is the API key used for authentication in the Alby platform.
	APIKey string
}

// NewAlbyClient returns a new client for the Alby wallet API
func NewAlbyClient(apiKey string) *AlbyClient {
	return &AlbyClient{
		APIKey: apiKey,
	}
}

// AlbyPaymentRequest is the request body for the Alby payment endpoint.
type AlbyPaymentRequest struct {
	Invoice string `json:"invoice"`
}

// AlbyPaymentResponse is the response body for the Alby payment endpoint.
type AlbyPaymentResponse struct {
	PaymentPreimage string `json:"payment_preimage"`
}

// GetPreimage returns the preimage for the given LN invoice.
func (a *AlbyClient) GetPreimage(invoice string) (string, error) {
	// Get the payment bolt 11 endpoint URL.
	url := fmt.Sprintf("%s/payments/bolt11", albyURL)

	// Create the request body.
	body := AlbyPaymentRequest{
		Invoice: invoice,
	}

	// Convert the request body to JSON.
	reqBodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("unable to encode request body: %w", err)
	}

	// Create the request.
	req, err := http.NewRequest(
		http.MethodPost, url, bytes.NewBuffer(reqBodyBytes),
	)
	if err != nil {
		return "", fmt.Errorf("unable to create request: %w", err)
	}

	// Set the Authorization header
	req.Header.Set("Authorization", "Bearer "+a.APIKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Send the request.
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
		return "", fmt.Errorf("unexpected response(%d): %s", resp.StatusCode,
			respBodyBytes)
	}

	var paymentResponse AlbyPaymentResponse
	err = json.Unmarshal(respBodyBytes, &paymentResponse)
	if err != nil {
		return "", fmt.Errorf("unable to parse response body: %w", err)
	}

	return paymentResponse.PaymentPreimage, nil
}
