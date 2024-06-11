package users

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/fewsats/fewsatscli/client"
	"github.com/urfave/cli/v2"
)

var updateUserDetailsCommand = &cli.Command{
	Name:   "update-details",
	Usage:  "Update user details",
	Action: updateUserDetails,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "username",
			Usage:    "New username of the user",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "profile-image",
			Usage:    "File path of the profile image to upload",
			Required: true,
		},
	},
}

func updateUserDetails(c *cli.Context) error {
	username := c.String("username")
	profileImagePath := c.String("profile-image")
	// Create a new multipart writer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add username field
	err := writer.WriteField("username", username)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to write username field: %s", err), 1)
	}

	// Handle the cover image
	coverFile, err := os.Open(profileImagePath)
	if err != nil {
		return cli.Exit("failed to open cover image file", 1)
	}
	defer coverFile.Close()

	// Read the entire file into memory
	coverData, err := io.ReadAll(coverFile)
	if err != nil {
		return cli.Exit("failed to read cover image file", 1)
	}

	// Encode to base64
	base64CoverData := base64.StdEncoding.EncodeToString(coverData)

	// Write base64 encoded data as a form field
	err = writer.WriteField("profile_image", base64CoverData)
	if err != nil {
		slog.Debug(
			"Failed to write cover field.",
			"error", err,
		)
		return cli.Exit("failed to write cover field", 1)
	}

	// Close the writer to finalize the multipart message
	err = writer.Close()
	if err != nil {
		return cli.Exit("failed to finalize multipart message", 1)
	}

	// Create a new HTTP client and request
	client, err := client.NewHTTPClient()
	if err != nil {
		slog.Debug("Failed to create HTTP client.", "error", err)
		return cli.Exit(fmt.Sprintf("Failed to create HTTP client: %s", err), 1)
	}

	// Execute the request
	resp, err := client.ExecuteMultipartRequest(
		http.MethodPut, "/v0/users/details",
		body.Bytes(), writer.FormDataContentType(),
	)
	if err != nil {
		slog.Debug("Failed to update user details.", "error", err)
		return cli.Exit(fmt.Sprintf("Failed to update user details: %s", err), 1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Debug(
			"Failed to update user details with status code.", "status_code", resp.StatusCode,
		)
		return cli.Exit(fmt.Sprintf("Failed to update user details with status code: %d", resp.StatusCode), 1)
	}

	fmt.Println("User details updated successfully.")
	return nil
}
