// Implements the 'q' command for asking queries to the selected model.
// It handles query formatting, file input, piped input, and sends requests to the Ollama API.

package chat

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/thejasmeetsingh/oclai/pkg/app"
	"github.com/thejasmeetsingh/oclai/pkg/config"
)

var fileContents []string

// Query is a subcommand for asking a query to the selected model.
var Query = &cobra.Command{
	Use:     "query [query]",
	Aliases: []string{"q"},
	Short:   "Ask a query to the model",
	Long:    color.New(color.FgBlue, color.Bold).Sprint("Ask a query to the model. You can provide a query directly, pipe input from another command, or specify a file to analyze."),
	Args:    cobra.MinimumNArgs(1),
	Example: `
		oclai query "Hey what's up --model qwen3:latest"
		cat /path/file.txt | oclai q "Summerize this file"
		oclai q "Analyze this code" -f /path/main.py
	`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return config.DefaultModelCheck()
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Join the query arguments into a single string.
		query := strings.TrimSpace(strings.Join(args, " "))
		errMsg := color.New(color.FgRed)

		if query == "" {
			errMsg.Println("Please provide a query 😒")
			return
		}

		// Check for piped input if fileContents is empty
		if len(fileContents) == 0 {
			err := readPipedInput()
			if err != nil {
				errMsg.Println(err.Error())
				os.Exit(1)
			}
		}

		// Format query if fileContents is not empty
		if len(fileContents) != 0 {
			query = fmt.Sprintf("```\n%s\n```\nUser Query: %s", strings.Join(fileContents, "\n"), query)
		}

		request := app.ModelRequest{
			Model: config.OclaiConfig.DefaultModel,
			Think: false,
			Messages: &[]app.Message{{
				Role:    app.User,
				Content: query,
			}},
		}

		// Send a one-off chat request to Ollama API
		content, err := app.Chat(request, true)
		if err != nil {
			errMsg.Println(err)
			os.Exit(1)
		}

		fmt.Println(content)
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
		return fmt.Errorf("failed to open file: %w", err)
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
