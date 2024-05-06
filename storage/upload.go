package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/fewsats/fewsatscli/client"
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
	FileURL      string `json:"file_url"`
}

// MultipartUploadRequest is adjusted to include multipart details
type MultipartUploadRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	PriceInCents uint64 `json:"price_in_cents"`
	NumParts     int    `json:"num_parts"`
	PartSize     int    `json:"part_size"`
}

// UploadFileResponse is the response body for the upload endpoint.
type UploadFileResponse struct {
	UploadID      string   `json:"upload_id"`
	PresignedURLs []string `json:"presigned_urls"`
}

type MultipartUploadInitResponse struct {
	UploadID      string   `json:"upload_id"`
	PresignedURLs []string `json:"presigned_urls"`
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

// uploadFile uploads a file to the storage service based on the input type.
func uploadFile(c *cli.Context) error {
	err := client.RequiresLogin()
	if err != nil {
		return cli.Exit("You need to log in to run this command.", 1)
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

	if filePath != "" {
		return handleMultipartUpload(c, filePath, name, description, price)
	} else {
		return cli.Exit("file-url upload not supported yet", 1)
		// return handleServerSideUpload(c, fileURL, name, description, price)
	}
}

// handleMultipartUpload handles the multipart upload of a local file.
func handleMultipartUpload(c *cli.Context, filePath, name, description string, price float64) error {
	numParts, partSize, err := calculateUploadParts(filePath)
	if err != nil {
		slog.Debug("Failed to prepare multipart upload data", "error", err)
		return cli.Exit("failed to prepare multipart upload data", 1)
	}

	// Prepare the request to initiate multipart upload
	uploadData := MultipartUploadRequest{
		Name:         name,
		Description:  description,
		PriceInCents: uint64(math.Floor(price * 100)),
		NumParts:     numParts,
		PartSize:     partSize,
	}

	reqBody, err := json.Marshal(uploadData)
	if err != nil {
		slog.Debug("Failed to marshal upload data", "error", err)
		return cli.Exit("failed to marshal upload data", 1)
	}

	// Send the initiation request
	httpClient, err := client.NewHTTPClient()
	if err != nil {
		slog.Debug("Failed to create HTTP client", "error", err)
		return cli.Exit("failed to create HTTP client", 1)
	}

	resp, err := httpClient.ExecuteRequest(http.MethodPost, uploadFilePath, reqBody)
	if err != nil || resp.StatusCode != http.StatusOK {
		slog.Debug("Failed to initiate multipart upload", "error", err, "statusCode", resp.StatusCode)
		return cli.Exit("failed to initiate multipart upload", 1)
	}

	var initResp MultipartUploadInitResponse
	err = json.NewDecoder(resp.Body).Decode(&initResp)
	if err != nil {
		slog.Debug("Failed to decode initiation response", "error", err)
		return cli.Exit("failed to decode initiation response", 1)
	}

	fmt.Printf("initResp: %+v\n", initResp)
	// Example usage of the new response structure
	if len(initResp.PresignedURLs) == 0 {
		return cli.Exit("no presigned URLs received", 1)
	}

	// Upload each part
	file, err := os.Open(filePath)
	if err != nil {
		slog.Debug("Failed to open file for reading", "error", err)
		return cli.Exit("failed to open file for reading", 1)
	}
	defer file.Close()
	var etags []string

	buffer := make([]byte, partSize)
	for i, url := range initResp.PresignedURLs {
		bytesRead, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			slog.Debug("Failed to read file part", "error", err)
			return cli.Exit("failed to read file part", 1)
		}

		partURL := url + "&partNumber=" + strconv.Itoa(i) + "&uploadId=" + initResp.UploadID
		req, err := http.NewRequest("PUT", partURL, bytes.NewReader(buffer[:bytesRead]))
		if err != nil {
			slog.Debug("Failed to create request for part upload", "error", err)
			return cli.Exit("failed to create request for part upload", 1)
		}

		partResp, err := http.DefaultClient.Do(req)
		if err != nil || partResp.StatusCode != http.StatusOK {
			slog.Debug("Failed to upload part", "error", err, "statusCode", partResp.StatusCode)
			return cli.Exit("failed to upload part", 1)
		}
		etags = append(etags, partResp.Header.Get("ETag"))

	}

	// Notify server of completion
	completeURL := initResp.PresignedURLs[0] + "?uploadId=" + initResp.UploadID
	completeReq := struct {
		Parts []struct {
			ETag       string `json:"etag"`
			PartNumber int    `json:"part_number"`
		} `json:"parts"`
	}{
		Parts: make([]struct {
			ETag       string `json:"etag"`
			PartNumber int    `json:"part_number"`
		}, len(etags)),
	}
	for i, etag := range etags {
		completeReq.Parts[i].ETag = etag
		completeReq.Parts[i].PartNumber = i + 1
	}

	reqBody, err = json.Marshal(completeReq)
	if err != nil {
		slog.Debug("Failed to marshal completion request", "error", err)
		return cli.Exit("failed to marshal completion request", 1)
	}

	req, err := http.NewRequest("POST", completeURL, bytes.NewBuffer(reqBody))
	if err != nil {
		slog.Debug("Failed to create completion request", "error", err)
		return cli.Exit("failed to create completion request", 1)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		slog.Debug("Failed to complete multipart upload", "error", err, "statusCode", resp.StatusCode)
		return cli.Exit("failed to complete multipart upload", 1)
	}

	fmt.Println("File uploaded successfully")
	return nil
}

// prepareMultipartUploadData prepares the necessary data for a multipart upload request.
func calculateUploadParts(filePath string) (int, int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		slog.Debug("Failed to open file", "error", err)
		return 0, 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		slog.Debug("Failed to get file info", "error", err)
		return 0, 0, fmt.Errorf("failed to get file info: %w", err)
	}

	const partSize = 125 * 1024 * 1024 // 125MB
	numParts := int(math.Ceil(float64(fileInfo.Size()) / float64(partSize)))

	return numParts, partSize, nil
	// return &UploadFileRequest{
	// 	Name:         filepath.Base(filePath),
	// 	Description:  "Description of the file", // This should be dynamically set based on context or user input
	// 	PriceInCents: 100,                       // This should be dynamically set based on context or user input
	// 	NumParts:     numParts,
	// 	PartSize:     partSize,
	// }, nil
}
