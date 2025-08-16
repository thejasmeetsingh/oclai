package query

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
)

var fileContents []string

// query is a subcommand for asking a query to the selected model.
var Query = &cobra.Command{
	Use:   "q [query]",
	Short: "Ask a query",
	Long:  "Ask a query to the model",
	Args:  cobra.MinimumNArgs(1),
	Example: `
		oclai q "Hey what's up"
		cat /path/file.txt | oclai q "Summerize this file"
		oclai q "Analyze this code" -f /path/main.py
	`,
	PreRunE: func(cmd *cobra.Command, args []string) error { return config.DefaultModelCheck() },
	Run: func(cmd *cobra.Command, args []string) {
		// Join the query arguments into a single string.
		query := strings.Join(args, " ")

		if query == "" {
			return
		}

		// Check for piped input if fileContents is empty
		if len(fileContents) == 0 {
			err := readPipedInput()
			if err != nil {
				fmt.Printf("😬 Error while reading the piped input %s", err.Error())
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

func isValidFilePath(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

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

func readFileContent(filePath string) error {
	if !isValidFilePath(filePath) {
		return fmt.Errorf("%s is not a valid file path ☹️", filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return readFromReader(file)
}

func readPipedInput() error {
	// Check if we have piped input
	stat, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("❌ failed to stat stdin: %w", err)
	}

	// If data is being piped in
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		return readFromReader(os.Stdin)
	}

	return nil
}

func init() {
	Query.PersistentFlags().FuncP("file", "f", "Read from file and ask query about the content", readFileContent)
}
