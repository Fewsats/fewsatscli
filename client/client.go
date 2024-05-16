package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/fewsats/fewsatscli/config"
	"github.com/fewsats/fewsatscli/credentials"
	"github.com/fewsats/fewsatscli/store"
	"github.com/fewsats/fewsatscli/wallets"
	"github.com/lightningnetwork/lnd/zpay32"
)

// HttpClient is an HTTP client for interacting with the Fewsats API.
type HttpClient struct {
	// client is the HTTP client used to make requests.
	client *http.Client

	// wallet is the interface provider used to get the preimage of an invoice.
	wallet wallets.PreimageProvider

	// apiKey is the API key used for authentication in our platform.
	apiKey        string
	domain        string
	albyToken     string
	sessionCookie *http.Cookie
}

// NewHTTPClient creates a new HTTP client for interacting with the Fewsats API.
func NewHTTPClient() (*HttpClient, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to create http client: %w", err)
	}

	store := store.GetStore()
	apiKey, err := store.GetAPIKey()
	if err != nil {
		return nil, fmt.Errorf("unable to get valid API key: %w", err)
	}

	wallet, err := wallets.GetDefaultWallet(store)
	switch {
	case errors.Is(err, wallets.ErrNoWalletFound):
		// No wallet found, continue without wallet.
	case err != nil:
		return nil, fmt.Errorf("unable to get default wallet: %w", err)
	}

	return &HttpClient{
		client:    &http.Client{},
		wallet:    wallet,
		apiKey:    apiKey,
		domain:    cfg.Domain,
		albyToken: cfg.AlbyToken,
	}, nil
}

func (c *HttpClient) SetSessionCookie(sessionCookie *http.Cookie) {
	c.sessionCookie = sessionCookie
}

func (c *HttpClient) ExecuteRequest(method, path string,
	body []byte) (*http.Response, error) {

	url := fmt.Sprintf("%s%s", c.domain, path)
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	// do not require auth for signup / login
	requireAuth := !strings.Contains(path, "/signup") &&
		!strings.Contains(path, "/login")

	switch {
	case c.apiKey != "":
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	case c.sessionCookie != nil:
		req.AddCookie(c.sessionCookie)
	case requireAuth:
		// this should not happen, because we use RequiresLogin() to check
		// if the user is logged in beforehand
		return nil, fmt.Errorf("you need to log in to run this command")
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

// getExternalID retrieves the external ID from the URL.
func getExternalID(url string) string {
	urlParts := strings.Split(url, "/")
	return urlParts[len(urlParts)-1]

}

// getL402Credentials retrieves the L402 credentials from the database.
func getL402Credentials(externalID string) (*credentials.L402Credentials,
	error) {

	store := store.GetStore()
	creds, err := store.GetL402Credentials(externalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get L402 credentials from db: %w",
			err)
	}

	return creds, nil
}

func saveL402Credentials(creds *credentials.L402Credentials) error {
	store := store.GetStore()
	err := store.InsertL402Credentials(creds)
	if err != nil {
		return fmt.Errorf("failed to insert credentials to db: %w", err)
	}

	return nil
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

	// check if we already paid the invoice and it's in the DB
	externalID := getExternalID(url)
	creds, err := getL402Credentials(externalID)

	switch {
	case errors.Is(err, credentials.ErrNoCredentialsFound):
		// No credentials found, continue without credentials.
	case err != nil:
		return nil, fmt.Errorf("unable to get L402 credentials: %w", err)
	default:
		slog.Debug(
			"Using existing L402 credentials",
			"macaroon", creds.Macaroon,
			"preimage", creds.Preimage,
		)

		header, err := creds.AuthenticationHeader()
		if err != nil {
			return nil, fmt.Errorf("unable to generate L402 auth header: %w",
				err)
		}

		req.Header.Set("Authorization", header)
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

	// Make sure we have a valid wallet to pay the invoice.
	if c.wallet == nil {
		fmt.Println()
		fmt.Println("unable to access L402 paywalled content: no wallet configured")
		fmt.Println("run `fewsatscli wallet connect` to connect your wallet")
		fmt.Println()

		return nil, fmt.Errorf("unable to access L402 paywalled content")
	}

	creds, err = credentials.ParseL402Challenge(externalID, resp)
	if err != nil {
		return nil, fmt.Errorf("unable to parse L402 challenge: %w", err)
	}

	invoicePrice, err := DecodePrice(creds.Invoice)
	if err != nil {
		return nil, fmt.Errorf("unable to decode invoice price: %w", err)
	}

	fmt.Printf("URL: %s\n", url)
	fmt.Printf("Lightning invoice price: %d sats\n", invoicePrice)
	fmt.Print("Do you want to continue? (y/N): ")

	var input string
	_, err = fmt.Scanln(&input)
	if err != nil {
		return nil, fmt.Errorf("unable to read user input: %w", err)
	}

	input = strings.TrimSpace(input)
	if input != "Y" && input != "y" {
		return nil, fmt.Errorf("user chose not to continue")
	}

	preimage, err := c.wallet.GetPreimage(creds.Invoice)
	if err != nil {
		return nil, fmt.Errorf("unable to pay invoice: %w", err)
	}

	creds.Preimage = preimage

	slog.Debug(
		"Paid invoice",
		"macaroon", creds.Macaroon,
		"invoice", creds.Invoice,
		"preimage", creds.Preimage,
	)

	err = saveL402Credentials(creds)
	if err != nil {
		return nil, fmt.Errorf("unable to save L402 credentials: %w", err)
	}

	authHeader, err := creds.AuthenticationHeader()
	if err != nil {
		return nil, fmt.Errorf("unable to generate L402 auth header: %w", err)
	}

	req.Header.Set("Authorization", authHeader)
	resp, err = c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to execute request: %w", err)
	}

	return resp, nil
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

// RequiresLogin checks for a valid API key and verifies it against
// the /authorize endpoint.
func RequiresLogin() error {
	store := store.GetStore()
	apiKeys, err := store.GetEnabledAPIKeys()
	if err != nil {
		return fmt.Errorf("failed to retrieve API keys: %w", err)
	}

	for _, apiKey := range apiKeys {
		resp, err := verifyAPIKey(apiKey.Key)
		if err != nil {
			return fmt.Errorf("failed to verify API key: %w", err)
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			return nil
		}

		if resp.StatusCode == http.StatusUnauthorized {
			store.DisableAPIKey(apiKey.ID)
		}
	}

	return fmt.Errorf("no valid API keys found")
}

// verifyAPIKey makes a request to the /authorize endpoint to check if the API
// key is still valid.
func verifyAPIKey(key string) (*http.Response, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to create http client: %w", err)
	}

	// using list apikeys to verify, ideally change to a custom authorize endpoint
	url := fmt.Sprintf("%s%s", cfg.Domain, "/v0/auth/apikeys")

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		slog.Debug("Failed to create request", "error", err, "key", key)
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Debug("Failed to execute authorize request", "error", err,
			"key", key)

		return nil, err
	}

	return resp, nil
}
