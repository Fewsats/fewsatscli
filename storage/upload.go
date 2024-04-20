package storage

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/fewsats/fewsatscli/client"
	"github.com/fewsats/fewsatscli/config"
	"github.com/urfave/cli/v2"
)

const (
	// uploadFilePath is the path to the upload endpoint.
	uploadFilePath = "/v0/storage/upload"
)

// UploadFileRequest is the request body for the upload endpoint.
type UploadFileRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	PriceInCents uint64 `json:"price_in_usd_cents"`
	File         []byte `json:"file"`
	FileURL      string `json:"file_url"`
}

// UploadFileResponse is the response body for the upload endpoint.
type UploadFileResponse struct {
	FileID string `json:"file_id"`
}

var uploadFileCommand = &cli.Command{
	Name:  "upload",
	Usage: "Upload a file to the storage service.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "name",
			Usage: "The name of the file stored in the storage service.",
		},
		&cli.StringFlag{
			Name:     "description",
			Usage:    "The description of the file contents.",
			Required: true,
		},
		&cli.StringFlag{
			Name:        "price",
			Usage:       "The price of the file in USD.",
			Required:    true,
			DefaultText: "19.99",
		},
		&cli.StringFlag{
			Name:  "file-path",
			Usage: "The file to upload.",
		},
		&cli.StringFlag{
			Name:  "file-url",
			Usage: "The URL where the file is stored.",
		},
	},
	Action: uploadFile,
}

// uploadFile uploads a file to the storage service.
func uploadFile(c *cli.Context) error {
	cfg, err := config.GetConfig()
	if err != nil {
		slog.Debug(
			"Failed to get config.",
			"error", err,
		)

		return cli.Exit("failed to get config", 1)
	}

	name := c.String("name")
	description := c.String("description")
	priceStr := c.String("price")
	filePath := c.String("file-path")
	fileURL := c.String("file-url")

	if filePath == "" && fileURL == "" {
		return cli.Exit("file-path or file-url is required", 1)
	}

	if filePath != "" && fileURL != "" {
		return cli.Exit("only one of file-path or file-url is allowed", 1)
	}

	if name == "" {
		if filePath != "" {
			name = filepath.Base(filePath)
		} else {
			return cli.Exit("name is required", 1)
		}
	}

	if description == "" {
		return cli.Exit("description is required", 1)
	}

	if priceStr == "" {
		return cli.Exit("price is required", 1)
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return cli.Exit("price must be a number (ex: 10.95)", 1)
	}

	var file []byte
	if filePath != "" {
		file, err = os.ReadFile(filePath)
		if err != nil {
			slog.Debug(
				"Failed to read file.",
				"error", err,
			)

			return cli.Exit("failed to read file", 1)
		}
	}

	req := &UploadFileRequest{
		Name:         name,
		Description:  description,
		PriceInCents: uint64(math.Floor(price * 100)),
		File:         file,
		FileURL:      fileURL,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		slog.Debug(
			"Failed to marshal request.",
			"error", err,
		)

		return cli.Exit("failed to marshal request", 1)
	}

	client, err := client.NewHTTPClient()
	if err != nil {
		slog.Debug(
			"Failed to create HTTP client.",
			"error", err,
		)

		return cli.Exit("failed to create HTTP client", 1)
	}

	method := http.MethodPost
	resp, err := client.ExecuteRequest(method, uploadFilePath, reqBody)
	if err != nil {
		slog.Debug(
			"Failed to execute request.",
			"error", err,
		)

		return cli.Exit("failed to execute request", 1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Debug(
			"Request failed",
			"status_code", resp.StatusCode,
		)

		return cli.Exit("failed to upload file", 1)
	}

	var respBody UploadFileResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		slog.Debug(
			"Failed to decode response.",
			"error", err,
		)

		return cli.Exit("failed to decode response", 1)
	}

	fmt.Println("File uploaded successfully.")
	fmt.Println("Download URL: ", cfg.Domain+downloadFilePath+"/"+respBody.FileID)

	return nil
}
