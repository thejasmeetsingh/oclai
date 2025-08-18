// Implements the 'q' command for asking queries to the selected model.
// It handles query formatting, file input, piped input, and sends requests to the Ollama API.

package chat

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thejasmeetsingh/oclai/pkg/config"
	"github.com/thejasmeetsingh/oclai/pkg/markdown"
)

var fileContents []string

// Query is a subcommand for asking a query to the selected model.
var Query = &cobra.Command{
	Use:   "q [query]",
	Short: "Ask a query to the model",
	Long:  "Ask a query to the model. You can provide a query directly, pipe input from another command, or specify a file to analyze.",
	Args:  cobra.MinimumNArgs(1),
	Example: `
		oclai q "Hey what's up"
		cat /path/file.txt | oclai q "Summerize this file"
		oclai q "Analyze this code" -f /path/main.py
	`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return config.DefaultModelCheck()
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Join the query arguments into a single string.
		query := strings.Join(args, " ")

		if query == "" {
			fmt.Println("Please provide a query 😒")
			return
		}

		// Check for piped input if fileContents is empty
		if len(fileContents) == 0 {
			err := readPipedInput()
			if err != nil {
				fmt.Printf("❌ Error while reading the piped input: %s\n", err.Error())
			}
		}

		// Format query if fileContents is not empty
		if len(fileContents) != 0 {
			query = fmt.Sprintf("```\n%s\n```\nUser Query: %s", strings.Join(fileContents, "\n"), query)
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
			fmt.Printf("❌ Error while generating a response: %s\n", err.Error())
			os.Exit(1)
		}

		// Check if the response status is OK.
		if response.StatusCode != http.StatusOK {
			fmt.Printf("❌ Received invalid response from Ollama service - Status Code: %d\n", response.StatusCode)
			os.Exit(1)
		}

		// Decode the response into the ModelResponse struct.
		type ModelResponse struct {
			Response string `json:"response"` // The response from the AI model.
		}

		var modelResponse ModelResponse
		if err = json.NewDecoder(response.Body).Decode(&modelResponse); err != nil {
			fmt.Printf("😬 Error while parsing the model response: %s\n", err.Error())
			os.Exit(1)
		}

		// Render the markdown response.
		if err = markdown.Render(modelResponse.Response); err != nil {
			fmt.Printf("😶 Error caught while rendering response: %s\n", err.Error())
			os.Exit(1)
		}
	},
}

// isValidFilePath checks if a file exists at the given path.
func isValidFilePath(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// readFromReader reads content from a file reader and stores it in fileContents.
func readFromReader(reader *os.File) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fileContents = append(fileContents, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// readFileContent reads content from a file and stores it in fileContents.
func readFileContent(filePath string) error {
	if !isValidFilePath(filePath) {
		return fmt.Errorf("'%s' is not a valid file path", filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("☹️ failed to open file: %w", err)
	}
	defer file.Close()

	return readFromReader(file)
}

// readPipedInput reads content from standard input if it's being piped.
func readPipedInput() error {
	// Check if we have piped input
	stat, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("failed to check stdin status: %w", err)
	}

	// If data is being piped in
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		return readFromReader(os.Stdin)
	}

	return nil
}

// init adds the file flag to the Query command.
func init() {
	Query.PersistentFlags().FuncP("file", "f", "Read from file and ask query about the content", readFileContent)
}
