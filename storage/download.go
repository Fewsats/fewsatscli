package storage

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/fewsats/fewsatscli/client"
	"github.com/fewsats/fewsatscli/config"
	"github.com/urfave/cli/v2"
)

const (
	// downloadFilePath is the path to the download endpoint.
	downloadFilePath = "/v0/storage/download"
)

var downloadFileCommand = &cli.Command{
	Name:      "download",
	Usage:     "Download a file from the storage service.",
	ArgsUsage: "<file_id>",
	Action:    downloadFile,
}

// downloadFile downloads a file from the storage service.
func downloadFile(c *cli.Context) error {
	cfg, err := config.GetConfig()
	if err != nil {
		return cli.Exit("failed to get config", 1)
	}

	if c.Args().Len() < 1 {
		return cli.Exit("missing <file_id> argument", 1)
	}

	file_url := c.Args().Get(0)
	// Check if the file_url is a valid URL
	// If not, it means that the user used the file_id, prepend the domain to
	// create a valid URL
	_, err = url.ParseRequestURI(file_url)
	if err != nil {
		file_url = cfg.Domain + downloadFilePath + "/" + file_url
	}

	httpClient, err := client.NewHTTPClient()
	if err != nil {
		slog.Debug(
			"Failed to create http client.",
			"error", err,
		)

		return cli.Exit("failed to create http client", 1)
	}

	resp, err := httpClient.ExecuteL402Request(http.MethodGet, file_url, nil, nil)
	if err != nil {
		slog.Debug(
			"Failed to execute L402 request.",
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

		return cli.Exit("failed to download file", 1)
	}

	fileName := resp.Header.Get("file-name")
	if fileName == "" {
		slog.Debug(
			"Failed to parse filename",
			"header", resp.Header,
		)

		return cli.Exit("failed to parse filename", 1)
	}

	// Create a new file
	outFile, err := os.Create(fileName)
	if err != nil {
		slog.Debug(
			"Failed to create file",
			"file_name", fileName,
			"error", err,
		)

		return cli.Exit("failed to create file", 1)
	}
	defer outFile.Close()

	// Copy the response body to the new file
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		slog.Debug(
			"Failed to write to file",
			"file_name", fileName,
			"error", err,
		)

		return cli.Exit("failed to write to file", 1)
	}

	fmt.Printf("File (%s) downloaded successfully.\n", fileName)

	return nil
}
