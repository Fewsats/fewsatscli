package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"mime/multipart"
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

// UploadFileResponse is the response body for the upload endpoint.
type UploadFileResponse struct {
	FileID       string `json:"file_id"`
	PresignedURL string `json:"presigned_url"`
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
		&cli.StringFlag{
			Name:  "cover-image",
			Usage: "The file path of the cover image to upload.",
		},
		&cli.StringSliceFlag{
			Name:  "tags",
			Usage: "Tags associated with the file.",
		},
	},
	Action: uploadFile,
}

// uploadFile uploads a file to the storage service.
func uploadFile(c *cli.Context) error {
	err := client.RequiresLogin()
	if err != nil {
		return cli.Exit("You need to log in to run this command.", 1)
	}

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
	coverImagePath := c.String("cover-image")

	if filePath == "" {
		return cli.Exit("file-path is required", 1)
	}

	if name == "" {
		name = filepath.Base(filePath)
	}

	if name == "" && filePath == "" {
		return cli.Exit("name is required", 1)
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

	priceInCents := uint64(math.Floor(price * 100))

	var file *os.File
	if filePath != "" {
		file, err = os.Open(filePath)
		if err != nil {
			slog.Debug(
				"Failed to read file.",
				"error", err,
			)

			return cli.Exit("failed to read file", 1)
		}
		defer file.Close()
	}

	// Create a new multipart writer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add existing fields
	err = writer.WriteField("name", name)
	if err != nil {
		slog.Debug(
			"Failed to write name field.",
			"error", err,
		)
		return cli.Exit("failed to write name field", 1)
	}

	// Add existing fields
	err = writer.WriteField("file_name", name)
	if err != nil {
		slog.Debug(
			"Failed to write file_name field.",
			"error", err,
		)
		return cli.Exit("failed to write file_name field", 1)
	}

	err = writer.WriteField("description", description)
	if err != nil {
		slog.Debug(
			"Failed to write description field.",
			"error", err,
		)
		return cli.Exit("failed to write description field", 1)
	}
	err = writer.WriteField("price_in_cents", strconv.FormatUint(priceInCents, 10))
	if err != nil {
		slog.Debug(
			"Failed to write price field.",
			"error", err,
		)
		return cli.Exit("failed to write price field", 1)
	}

	// Handle the cover image as a file instead of base64 string
	if coverImagePath != "" {
		coverFile, err := os.Open(coverImagePath)
		if err != nil {
			return cli.Exit("failed to open cover image file", 1)
		}
		defer coverFile.Close()

		part, err := writer.CreateFormFile("cover", filepath.Base(coverImagePath))
		if err != nil {
			return cli.Exit("failed to create form file for cover image", 1)
		}
		if _, err := io.Copy(part, coverFile); err != nil {
			return cli.Exit("failed to write cover image file to form", 1)
		}
	}

	// Handle tags
	tags := c.StringSlice("tags")
	for _, tag := range tags {
		if err := writer.WriteField("tags", tag); err != nil {
			return cli.Exit("failed to write tag field", 1)
		}
	}

	// Close the writer to finalize the multipart message
	err = writer.Close()
	if err != nil {
		return cli.Exit("failed to finalize multipart message", 1)
	}

	// Create a new HTTP client and request
	client, err := client.NewHTTPClient()
	if err != nil {
		slog.Debug(
			"Failed to create HTTP client.",
			"error", err,
		)

		return cli.Exit("failed to create HTTP client", 1)
	}

	// Execute the request
	resp, err := client.ExecuteMultipartRequest(
		http.MethodPost, uploadFilePath,
		body.Bytes(), writer.FormDataContentType(),
	)
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

	// presignedURL is empty when we are uploading via fileURL
	if respBody.PresignedURL != "" {
		err = uploadFileToPresignedURL(respBody.PresignedURL, file)
		if err != nil {
			return cli.Exit("failed to upload file to presigned URL", 1)
		}
	}

	fmt.Println("File uploaded successfully.")
	fmt.Println("Download URL: ", cfg.Domain+downloadFilePath+"/"+respBody.FileID)

	return nil
}

// uploadFileToPresignedURL uploads the file to the presigned URL.
func uploadFileToPresignedURL(presignedURL string, file *os.File) error {
	req, err := http.NewRequest(http.MethodPut, presignedURL, file)
	if err != nil {
		slog.Debug(
			"Failed to create request to presigned URL.",
			"error", err,
		)
		return err
	}

	stat, err := file.Stat()
	if err != nil {
		slog.Debug(
			"Failed to get file stats.",
			"error", err,
		)
		return err
	}

	req.ContentLength = stat.Size()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Debug(
			"Failed to upload file to presigned URL.",
			"error", err,
		)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Debug(
			"Upload to presigned URL failed.",
			"status_code", resp.StatusCode,
		)
		return fmt.Errorf("upload to presigned URL failed with status code %d", resp.StatusCode)
	}

	return nil
}
