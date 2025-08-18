// Defines the root command and initialization logic for the CLI application.
// It sets up persistent flags for configuring the base URL and default model,
// adds the query subcommand, and loads the configuration file during initialization.

package cmd

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thejasmeetsingh/oclai/pkg/chat"
	"github.com/thejasmeetsingh/oclai/pkg/config"
)

// rootCmd is the main command for the CLI application.
// It serves as the entry point for all CLI commands.
var rootCmd = &cobra.Command{
	Use:     "oclai",
	Short:   "A completely offline agentic CLI",
	Long:    "An offline agentic CLI that brings Claude Code and Gemini CLI capabilities to your terminal using local AI models.",
	Example: `oclai q "What's the latest news of today"`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 0 {
			return
		}

		// Check if any global flags have been changed
		globalCmds := []string{"baseURL", "model"}
		for _, gloglobalCmd := range globalCmds {
			if cmd.Flags().Lookup(gloglobalCmd).Changed {
				return
			}
		}

		// If no arguments were provided and no global flags were changed, show help
		cmd.Help()
	},
}

// setBaseURL configures the base URL for the Ollama API.
// It validates the input and updates the configuration if valid.
func setBaseURL(arg string) error {
	if arg == "" {
		return fmt.Errorf("❌ Base URL cannot be empty. Please provide a valid URL")
	}

	// Parse the URL and validate it
	baseURL, err := url.Parse(strings.TrimSpace(arg))
	if err != nil {
		return fmt.Errorf("❌ Invalid URL format: %w. Please enter a valid URL", err)
	}

	// Update the configuration with the new base URL
	config.OclaiConfig.BaseURL = baseURL.String()
	return config.UpdateConfig()
}

// setDefaultModel sets the default model to be used by the CLI.
// It validates the input and updates the configuration if valid.
func setDefaultModel(arg string) error {
	if arg == "" {
		return fmt.Errorf("❌ Model value cannot be empty. Please provide a valid model name")
	}

	// Update the configuration with the new default model
	config.OclaiConfig.DefaultModel = strings.TrimSpace(arg)
	return config.UpdateConfig()
}

// init initializes the root command by:
// 1. Adding persistent flags for baseURL and model
// 2. Adding the query subcommand
// 3. Loading the configuration file
func init() {
	// Add persistent flags to the root command
	rootCmd.PersistentFlags().Func("baseURL", "Set Ollama BaseURL", setBaseURL)
	rootCmd.PersistentFlags().Func("model", "Set Default Model", setDefaultModel)

	// Add the query subcommand
	rootCmd.AddCommand(chat.Query)

	// Load configuration file
	if err := config.LoadConfig(); err != nil {
		fmt.Println("❌ Failed to load configuration:", err)
		os.Exit(1)
	}
}

// Execute runs the root command.
// It handles any errors that occur during command execution.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println("❌ Error executing command:", err)
		os.Exit(1)
	}
}
