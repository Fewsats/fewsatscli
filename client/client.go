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
	"github.com/fewsats/fewsatscli/store"
	storePkg "github.com/fewsats/fewsatscli/store"
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
	sessionCookie *http.Cookie
}

// NewHTTPClient creates a new HTTP client for interacting with the Fewsats API.
func NewHTTPClient() (*HttpClient, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to create http client: %w", err)
	}

	store := storePkg.GetStore()
	apiKey, err := store.GetAPIKey()
	if err != nil {
		return nil, fmt.Errorf("unable to get valid API key: %w", err)
	}

	var preimageProvider wallets.PreimageProvider

	walletType, err := store.GetWalletType()
	if err != nil && !errors.Is(err, storePkg.ErrNoWalletFound) {
		return nil, fmt.Errorf("unable to get wallet type: %w", err)
	}

	switch walletType {
	case "alby":
		token, err := store.GetWalletToken()
		if err != nil {
			return nil, fmt.Errorf("unable to get wallet token: %w", err)
		}

		preimageProvider = NewAlbyClient(token)

	case "zbd":
		token, err := store.GetWalletToken()
		if err != nil {
			return nil, fmt.Errorf("unable to get wallet token: %w", err)
		}

		preimageProvider = NewZBDClient(token)

	default:
		return nil, fmt.Errorf("unsupported wallet type: %s", walletType)
	}

	return &HttpClient{
		client: &http.Client{},
		wallet: preimageProvider,

		apiKey: apiKey,
		domain: cfg.Domain,
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
func getL402Credentials(externalID string) (*store.L402Credentials, error) {
	store := store.GetStore()
	credentials, err := store.GetL402Credentials(externalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get L402 credentials from db: %w", err)
	}

	return credentials, nil
}

// saveL402Credentials saves the L402 credentials to the database.
func saveL402Credentials(credentials *store.L402Credentials) error {
	store := store.GetStore()
	err := store.InsertL402Credentials(credentials)
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
	credentials, err := getL402Credentials(externalID)
	if err != nil {
		slog.Debug(
			"No L402 credentials found",
			"error", err,
		)
	}

	if credentials != nil {
		slog.Debug(
			"Using existing L402 credentials",
			"macaroon", credentials.Macaroon,
			"preimage", credentials.Preimage,
		)

		req.Header.Set("Authorization", credentials.L402Header())
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

	if c.wallet == nil {
		return nil, fmt.Errorf("unable to access L402 paywalled content: run " +
			"`fewsatscli wallet connect` to connect your wallet")
	}

	credentials, err = store.ParseL402Challenge(externalID, resp)
	if err == nil {
		return nil, fmt.Errorf("unable to parse L402 challenge header: %w",
			err)
	}

	invoicePrice, err := DecodePrice(credentials.Macaroon)
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

	preimage, err := c.wallet.GetPreimage(credentials.Invoice)
	if err != nil {
		return nil, fmt.Errorf("unable to pay invoice: %w", err)
	}

	credentials.Preimage = preimage

	slog.Debug(
		"Paid invoice",
		"macaroon", credentials.Macaroon,
		"invoice", credentials.Invoice,
		"preimage", credentials.Preimage,
	)

	err = saveL402Credentials(credentials)
	if err != nil {
		return nil, fmt.Errorf("unable to save L402 credentials: %w", err)
	}

	req.Header.Set("Authorization", credentials.L402Header())
	resp, err = c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to execute request: %w", err)
	}

	return resp, nil
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
