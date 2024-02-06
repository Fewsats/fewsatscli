package recipes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/urfave/cli/v2"
)

var GetCodeRecipeCommand = &cli.Command{
	Name:  "get",
	Usage: "Retrive the details for the given code recipe.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "recipe_id",
			Usage: "The ID of the recipe to get details.",
		},
	},
	Action: getCodeRecipe,
}

type GetRecipeResponse struct {
	Description string `json:"description"`
	FilePath    string `json:"file_path"`
	Name        string `json:"name"`
	RecipeID    string `json:"recipe_id"`
	Verified    bool   `json:"verified"`
}

// getCodeRecipe gets the details of a code recipe.
func getCodeRecipe(c *cli.Context) error {
	recipeID := c.String("recipe_id")

	// If recipeID is empty, set it to the first argument from the command line
	if recipeID == "" {
		if c.NArg() > 0 {
			recipeID = c.Args().Get(0)
		} else {
			return fmt.Errorf("recipe_id is required")
		}
	}

	url := fmt.Sprintf("%s/recipes/%s", baseURL, recipeID)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var recipe GetRecipeResponse
	err = json.NewDecoder(resp.Body).Decode(&recipe)
	if err != nil {
		return err
	}

	// Print the recipe details line by line
	fmt.Println()
	fmt.Printf("Recipe ID: %s\n", recipe.RecipeID)
	fmt.Printf("Name: %s\n", recipe.Name)
	fmt.Printf("Description: %s\n", recipe.Description)
	fmt.Printf("Verified: %v\n", recipe.Verified)
	fmt.Println()

	return nil
}
