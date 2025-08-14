package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thejasmeetsingh/oclai/pkg/config"
)

var oclaiConfig config.Config

func loadConfig() error {
	file := config.GetConfigFile()
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &oclaiConfig)
}

func updateConfig() error {
	data, err := json.MarshalIndent(&oclaiConfig, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(oclaiConfig.File, data, 0644)
}

func preCheck() error {
	if oclaiConfig.DefaultModel == "" {
		return fmt.Errorf("please select a default model 🤖")
	}
	return nil
}

// rootCmd is the main command for the CLI.
var rootCmd = &cobra.Command{
	Use:     "oclai",
	Short:   "A completely offline agentic CLI",
	Long:    "An offline agentic CLI that brings Claude Code and Gemini CLI capabilities to your terminal using local AI models.",
	Example: `oclai q "What's the latest news of today"`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 0 {
			return
		}

		// Check if global flag values has changed or not
		isBaseURLChanged := cmd.Flags().Lookup("baseURL").Changed
		isDefaultModelChanged := cmd.Flags().Lookup("model").Changed

		if isBaseURLChanged || isDefaultModelChanged {
			return
		}

		cmd.Help()
	},
}

// chat is a subcommand for asking a query to the selected model.
var chat = &cobra.Command{
	Use:     "q [query]",
	Short:   "Ask a query",
	Long:    "Ask a query to the model",
	Args:    cobra.MinimumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error { return preCheck() },
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
			"model":  oclaiConfig.DefaultModel,
			"prompt": query,
			"think":  false,
		})

		// Send the request to the Ollama API.
		response, err := http.Post(oclaiConfig.BaseURL+"/api/generate", "application/json", body)
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

// Execute runs the root command.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// init adds the "chat" command as a subcommand to the root command.
func init() {
	rootCmd.PersistentFlags().Func("baseURL", "Set Ollama BaseURL", func(s string) error {
		if s == "" {
			return fmt.Errorf("❌ baseURL should not be empty")
		}

		baseURL, err := url.Parse(strings.TrimSpace(s))
		if err != nil {
			return fmt.Errorf("❌ please enter a valid URL")
		}

		oclaiConfig.BaseURL = baseURL.String()
		return updateConfig()
	})

	rootCmd.PersistentFlags().Func("model", "Set Default Model", func(s string) error {
		if s == "" {
			return fmt.Errorf("❌ model value should not be empty")
		}

		oclaiConfig.DefaultModel = strings.TrimSpace(s)
		return updateConfig()
	})

	rootCmd.AddCommand(chat)

	err := config.SetupConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err = loadConfig(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
