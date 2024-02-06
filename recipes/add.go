package recipes

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

var AddCodeRecipeCommand = &cli.Command{
	Name:  "add",
	Usage: "Upload a new code recipe.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "path",
			Usage:    "The path to the file or folder containing the code recipe.",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "name",
			Usage:    "The name of the code recipe.",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "description",
			Usage:    "The description of the code recipe.",
			Required: true,
		},
	},
	Action: addCodeRecipe,
}

func zipDir(src string) (*bytes.Buffer, error) {
	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)

	zipWriter := zip.NewWriter(buf)
	defer zipWriter.Close()

	// Walk through each file/directory in the source directory
	err := filepath.Walk(src, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create a header based on the fileInfo
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// If it's a file, set its name to its relative path
		// If it's a directory, add a trailing slash
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Name, err = filepath.Rel(src, filePath)
			if err != nil {
				return err
			}
		}

		// Create the file header in the zip
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// If it's a file, write its content to the zip
		if !info.IsDir() {
			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return buf, nil
}

// getRecipeZip returns a zip with all the related code to the recipe.
func getRecipeZip(path string) (*bytes.Buffer, error) {
	// Check that the path is correct
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)

	// If the path is a directory, make a zip of it in memory and return it
	if info.IsDir() {
		buf, err = zipDir(path)
		if err != nil {
			return nil, err
		}

		return buf, nil
	}

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// If the path is a zip file, we can just return it
	if strings.HasSuffix(info.Name(), ".zip") {
		_, err = buf.Write(data)
		if err != nil {
			return nil, err
		}

		return buf, nil
	}

	// Create a new zip file
	w := zip.NewWriter(buf)

	// Add the file to the zip
	f, err := w.Create(info.Name())
	if err != nil {
		return nil, err
	}
	_, err = f.Write(data)
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}

	return buf, nil
}

type AddCodeRecipeRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AddCodeRecipeResponse struct {
	RecipeID string `json:"recipe_id"`
	Verified bool   `json:"verified"`
}

// AddCodeRecipe adds a new code recipe.
func addCodeRecipe(c *cli.Context) error {
	path := c.String("path")
	name := c.String("name")
	description := c.String("description")

	// Get the zip file.
	zipFile, err := getRecipeZip(path)
	if err != nil {
		return err
	}

	// Create a buffer to hold the request body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create form
	part, err := writer.CreateFormFile("file", "code.zip")
	if err != nil {
		return err
	}

	// Copy the zip file to the form
	_, err = io.Copy(part, zipFile)
	if err != nil {
		return err
	}

	// Add the name and description as form fields
	err = writer.WriteField("name", name)
	if err != nil {
		return err
	}
	err = writer.WriteField("description", description)
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	// Make the POST request
	url := fmt.Sprintf("%s/recipes", baseURL)
	resp, err := http.Post(url, writer.FormDataContentType(), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	// Parse the response body
	var response AddCodeRecipeResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return err
	}

	// Print the response to the screen
	fmt.Println()
	fmt.Printf("Recipe ID: %s\n", response.RecipeID)
	fmt.Printf("Verified: %v\n", response.Verified)
	fmt.Println()

	return nil
}
