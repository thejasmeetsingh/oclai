// Defines the root command and initialization logic for the CLI application.
// It sets up persistent flags for configuring the base URL and default model,
// adds the query subcommand, and loads the configuration file during initialization.

package cmd

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/thejasmeetsingh/oclai/app"
	"github.com/thejasmeetsingh/oclai/app/chat"
	"github.com/thejasmeetsingh/oclai/ollama"
)

var (
	infoMsg    = color.New(color.FgBlue, color.Bold)
	errMsg     = color.New(color.FgRed)
	successMsg = color.New(color.FgGreen)
)

// rootCmd is the main command for the CLI application.
// It serves as the entry point for all CLI commands.
var rootCmd = &cobra.Command{
	Use:     "oclai",
	Short:   "A completely offline agentic CLI",
	Long:    infoMsg.Sprint("An offline agentic CLI that brings Claude Code and Gemini CLI capabilities to your terminal using local AI models."),
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

// Command for viewing ollama models
var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "List available models",
	Long:  infoMsg.Sprint("Display all models currently available in your local Ollama installation."),
	Run: func(cmd *cobra.Command, args []string) {
		content, err := ollama.ShowModels(app.OclaiConfig.BaseURL, nil)
		if err != nil {
			errMsg.Println("Error listing models:", err)
			os.Exit(1)
		}
		fmt.Println(content)
	},
}

// Command for checking ollama service status
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check Ollama service status",
	Long:  infoMsg.Sprint("Check if Ollama service is running and display connection information."),
	Run: func(cmd *cobra.Command, args []string) {
		if err := ollama.CheckOllamaConnection(app.OclaiConfig.BaseURL); err != nil {
			errMsg.Println("Ollama Status:", err)
			os.Exit(1)
		}

		successMsg.Println("✅ Ollama is running at:", app.OclaiConfig.BaseURL)
	},
}

// setBaseURL configures the base URL for the Ollama API.
// It validates the input and updates the configuration if valid.
func setBaseURL(arg string) error {
	if arg == "" {
		return fmt.Errorf("baseURL cannot be empty. Please provide a valid URL")
	}

	// Parse the URL and validate it
	baseURL, err := url.Parse(strings.TrimSpace(arg))
	if err != nil {
		return fmt.Errorf("invalid URL format: %w. Please enter a valid URL", err)
	}

	// Update the configuration with the new base URL
	app.OclaiConfig.BaseURL = baseURL.String()
	return app.UpdateConfig()
}

// setDefaultModel sets the default model to be used by the CLI.
// It validates the input and updates the configuration if valid.
func setDefaultModel(arg string) error {
	if arg == "" {
		return fmt.Errorf("model value cannot be empty. Please provide a valid model name")
	}

	// Update the configuration with the new default model
	app.OclaiConfig.DefaultModel = strings.TrimSpace(arg)
	return app.UpdateConfig()
}

// init initializes the root command by:
// - Adding persistent flags & subcommands.
// - Loading the configuration file
func init() {
	// Add persistent flags to the root command
	rootCmd.PersistentFlags().Func("baseURL", "Set Ollama BaseURL", setBaseURL)
	rootCmd.PersistentFlags().Func("model", "Set Default Model", setDefaultModel)

	// Add the query subcommand
	rootCmd.AddCommand(
		modelsCmd,
		statusCmd,
		chat.Query,
		chat.Chat,
	)

	// Load configuration file
	if err := app.LoadConfig(); err != nil {
		errMsg.Println("Failed to load configuration: ", err)
		os.Exit(1)
	}
}

// Execute runs the root command.
// It handles any errors that occur during command execution.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		errMsg.Println("Error executing command:", err)
		os.Exit(1)
	}
}
