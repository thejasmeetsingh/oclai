package query

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thejasmeetsingh/oclai/pkg/config"
)

// query is a subcommand for asking a query to the selected model.
var Query = &cobra.Command{
	Use:     "q [query]",
	Short:   "Ask a query",
	Long:    "Ask a query to the model",
	Args:    cobra.MinimumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error { return config.DefaultModelCheck() },
	Run: func(cmd *cobra.Command, args []string) {
		// Join the query arguments into a single string.
		query := strings.Join(args, " ")

		if query == "" {
			return
		}

		// Create a JSON body for the request.
		body := &bytes.Buffer{}
		encoder := json.NewEncoder(body)
		encoder.Encode(map[string]any{
			"model":  config.OclaiConfig.DefaultModel,
			"prompt": query,
			"think":  false,
		})

		// Send the request to the Ollama API.
		response, err := http.Post(config.OclaiConfig.BaseURL+"/api/generate", "application/json", body)
		if err != nil {
			fmt.Printf("❌ Error caught while generating a response: %s\n", err.Error())
			os.Exit(1)
		}

		// Check if the response status is OK.
		if response.StatusCode != http.StatusOK {
			fmt.Printf("❌ Received invalid response from Ollama service - StatusCode: %d\n", response.StatusCode)
			os.Exit(1)
		}

		// Decode the response into the ModelResponse struct.
		type ModelResponse struct {
			Response string `json:"response"` // The response from the AI model.
		}

		var modelResponse ModelResponse
		if err = json.NewDecoder(response.Body).Decode(&modelResponse); err != nil {
			fmt.Printf("😬 Error caught while parsing the model response: %s\n", err.Error())
			os.Exit(1)
		}

		// Print the response.
		fmt.Printf("* %s\n", modelResponse.Response)
	},
}
