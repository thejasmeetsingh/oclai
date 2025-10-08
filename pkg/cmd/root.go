package cmd

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thejasmeetsingh/oclai/pkg/app"
	"github.com/thejasmeetsingh/oclai/pkg/mcp"
	"github.com/thejasmeetsingh/oclai/pkg/ollama"
	"github.com/thejasmeetsingh/oclai/pkg/utils"
)

// rootPath stores the application root directory path
var rootPath = ""

var (
	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:     "oclai",
		Short:   "A completely offline agentic CLI",
		Long:    utils.InfoBox("An AI powered terminal assistant similar to Claude Code and Gemini CLI, but runs entirely offline using local models.\nNo API keys, no subscriptions, no data leaving your machine."),
		Example: `oclai q "Tell me about the roman empire"`,
		Run: func(cmd *cobra.Command, args []string) {
			// If there are arguments, do nothing (handled by other commands)
			if len(args) != 0 {
				return
			}

			// Check if any global flags have been changed
			globalCmds := []string{"baseURL", "model", "ctx"}
			for _, gloglobalCmd := range globalCmds {
				if cmd.Flags().Lookup(gloglobalCmd).Changed {
					return
				}
			}

			// If no flags changed, show help
			cmd.Help()
		},
	}

	// modelsCmd lists available models
	modelsCmd = &cobra.Command{
		Use:   "models",
		Short: "List available models",
		Long:  utils.InfoBox("Display all models currently available in your local Ollama installation."),
		Run: func(cmd *cobra.Command, args []string) {
			// Fetch and display models from Ollama
			content, err := ollama.ShowModels(app.OclaiConfig.BaseURL, nil)
			if err != nil {
				fmt.Println(utils.ErrorMessage(fmt.Sprintf("Error listing models: %s", err.Error())))
				os.Exit(1)
			}
			fmt.Println(content)
		},
	}

	// statusCmd checks Ollama service status
	statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Check Ollama service status",
		Long:  utils.InfoBox("Check if Ollama service is running and display connection information."),
		Run: func(cmd *cobra.Command, args []string) {
			// Check if Ollama is running
			if err := ollama.CheckOllamaConnection(app.OclaiConfig.BaseURL); err != nil {
				fmt.Println(utils.ErrorMessage(fmt.Sprintf("Ollama Status: %s", err.Error())))
				os.Exit(1)
			}

			fmt.Println(utils.SuccessBox(fmt.Sprintf("Ollama is running at: %s", app.OclaiConfig.BaseURL)))
		},
	}
)

// setBaseURL updates the base URL configuration
func setBaseURL(arg string) error {
	arg = strings.TrimSpace(arg)

	// Validate input
	if arg == "" {
		return fmt.Errorf("✗ 'baseURL' cannot be empty. Please provide a valid URL")
	}

	// Parse URL
	baseURL, err := url.Parse(strings.TrimSpace(arg))
	if err != nil {
		return fmt.Errorf("✗ invalid URL format: %s. Please enter a valid URL", err.Error())
	}

	// Update configuration
	app.OclaiConfig.BaseURL = baseURL.String()
	if err = app.UpdateConfig(rootPath); err != nil {
		return err
	}

	fmt.Println(utils.SuccessBox("BaseURL updated successfully!"))

	return nil
}

// setDefaultModel updates the default model configuration
func setDefaultModel(arg string) error {
	arg = strings.TrimSpace(arg)

	// Validate input
	if arg == "" {
		return fmt.Errorf("✗ model value cannot be empty. Please provide a valid model name")
	}

	model := strings.TrimSpace(arg)

	// Check if the model exists
	isExists, err := ollama.IsModelExists(app.OclaiConfig.BaseURL, model, nil)
	if err != nil {
		return err
	}

	if !isExists {
		return fmt.Errorf("'%s' model does not exists", model)
	}

	// Update configuration
	app.OclaiConfig.DefaultModel = strings.TrimSpace(arg)
	if err := app.UpdateConfig(rootPath); err != nil {
		return err
	}

	fmt.Println(utils.SuccessBox("Default Model updated successfully!"))

	return nil
}

// setNumCtx updates the context limit configuration
func setNumCtx(arg string) error {
	arg = strings.TrimSpace(arg)

	// Validate input
	if arg == "" {
		return fmt.Errorf("✗ 'num_ctx' value should not be empty")
	}

	// Convert to integer
	numCtx, err := strconv.Atoi(arg)
	if err != nil {
		return fmt.Errorf("✗ value should be a valid integer")
	}

	// Update configuration
	app.OclaiConfig.NumCtx = numCtx
	if err := app.UpdateConfig(rootPath); err != nil {
		return err
	}

	fmt.Println(utils.SuccessBox("NumCtx value updated successfully!"))

	return nil
}

// Retrieve the current version from the VERSION file
func getCurrVersion() string {
	data, err := utils.ReadConfig("VERSION")
	if err != nil {
		fmt.Println(utils.ErrorMessage("error caught while retrieving version information"))
		os.Exit(1)
	}

	version := strings.TrimSpace(string(data))
	return version
}

func init() {
	// Get application root directory
	_rootPath, err := utils.GetAppRootDir()
	if err != nil {
		fmt.Println(utils.ErrorMessage(fmt.Sprintf("Error caught while retreiving root path: %s", err.Error())))
		os.Exit(1)
	}

	rootPath = _rootPath

	// Register command flags
	rootCmd.PersistentFlags().Func("baseURL", "Set Ollama BaseURL", setBaseURL)
	rootCmd.PersistentFlags().Func("model", "Set Default Model", setDefaultModel)
	rootCmd.PersistentFlags().Func("ctx", "Set Context Limit", setNumCtx)

	// Add version information in root cmd
	rootCmd.Version = getCurrVersion()
	rootCmd.SetVersionTemplate(`Oclai version is {{printf "%s\n" .Version}}`)

	// Add sub-commands to root
	rootCmd.AddCommand(
		modelsCmd,
		statusCmd,
		app.Query,
		app.Chat,
		mcp.McpRootCmd,
	)
}

func Execute() {
	// Execute root command
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(utils.ErrorMessage(fmt.Sprintf("Error executing command: %s", err.Error())))
		os.Exit(1)
	}
}
