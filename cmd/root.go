package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// BASE_URL is the base URL for the Ollama API, read from the environment variable.
var BASE_URL = os.Getenv("BASE_URL")

// DEFAULT_MODEL is the default AI model used by the CLI.
var DEFAULT_MODEL = "qwen3:14b"

// Tags is a struct representing the response from the "/api/tags" endpoint.
// It contains a list of available models and their sizes.
type Tags struct {
	Models []struct {
		Model string `json:"model"` // Name of the model.
		Size  int    `json:"size"`  // Size of the model in bytes.
	} `json:"models"` // List of models.
}

// rootCmd is the main command for the CLI.
var rootCmd = &cobra.Command{
	Use:   "oclai",
	Short: "A completely offline agentic CLI",
	Long:  "An offline agentic CLI that brings Claude Code and Gemini CLI capabilities to your terminal using local AI models.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Fetch the list of available models from the Ollama API.
		response, err := http.Get(BASE_URL + "/api/tags")
		if err != nil || response.StatusCode != http.StatusOK {
			fmt.Printf("Cannot connect to Ollama: %s\n", err.Error())
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Fetch the list of available models from the Ollama API.
		response, err := http.Get(BASE_URL + "/api/tags")
		if err != nil {
			fmt.Printf("Error caught while fetching models: %s\n", err.Error())
			os.Exit(1)
		}
		defer response.Body.Close()

		// Decode the JSON response into the Tags struct.
		var tags Tags
		if err = json.NewDecoder(response.Body).Decode(&tags); err != nil {
			fmt.Printf("Error caught while parsing models data: %s\n", err.Error())
			os.Exit(1)
		}

		// Check if there are no available models.
		if len(tags.Models) == 0 {
			fmt.Println("No models available ☹️")
			os.Exit(1)
		}

		// Display the available models.
		fmt.Println("Available Models 🤖")
		fmt.Printf("%-4s %-10s %-50s\n", "-", "MODEL", "SIZE")
		fmt.Println(strings.Repeat("-", 50))

		for idx, model := range tags.Models {
			// Convert the model size from bytes to gigabytes.
			modelSize := fmt.Sprintf("%d GB", model.Size/int(math.Pow(1024, 3)))
			fmt.Printf("%-4d %-10s %-50s\n", idx+1, model.Model, modelSize)
		}

		// Prompt the user to select a default model.
		modelChoice := getModelChoice(tags)
		DEFAULT_MODEL = tags.Models[modelChoice].Model
	},
}

// chat is a subcommand for asking a query to the selected model.
var chat = &cobra.Command{
	Use:   "q [query]",
	Short: "Ask a query",
	Long:  "Ask a query to the selected model",
	Run: func(cmd *cobra.Command, args []string) {
		// Join the query arguments into a single string.
		query := strings.Join(args, " ")

		// Create a JSON body for the request.
		body := &bytes.Buffer{}
		encoder := json.NewEncoder(body)
		encoder.Encode(map[string]interface{}{
			"model":  DEFAULT_MODEL,
			"prompt": query,
			"think":  false,
		})

		// Send the request to the Ollama API.
		response, err := http.Post(BASE_URL+"/api/generate", "application/json", body)
		if err != nil {
			fmt.Printf("Error caught while generating a response: %s\n", err.Error())
			os.Exit(1)
		}

		// Check if the response status is OK.
		if response.StatusCode != http.StatusOK {
			fmt.Printf("Received invalid response from Ollama service - StatusCode: %d\n", response.StatusCode)
			os.Exit(1)
		}

		// Decode the response into the ModelResponse struct.
		type ModelResponse struct {
			Response string `json:"response"` // The response from the AI model.
		}
		var modelResponse ModelResponse
		if err = json.NewDecoder(response.Body).Decode(&modelResponse); err != nil {
			fmt.Printf("Error caught while parsing the model response: %s\n", err.Error())
			os.Exit(1)
		}

		// Print the response.
		fmt.Println(modelResponse.Response)
	},
}

// getModelChoice prompts the user to select a model from the list of available models.
func getModelChoice(tags Tags) int {
	fmt.Println("Select a default model:")

	var modelChoice int
	_, err := fmt.Scanf("%d", &modelChoice)
	if err != nil {
		fmt.Println("Please enter a valid integer")
		return getModelChoice(tags)
	}

	// Validate the user's input.
	if modelChoice < 1 || modelChoice > (len(tags.Models)+1) {
		fmt.Println("Invalid choice")
		return getModelChoice(tags)
	}

	return modelChoice
}

// Execute runs the root command.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// init adds the "chat" command as a subcommand to the root command.
func init() {
	rootCmd.AddCommand(chat)
}
