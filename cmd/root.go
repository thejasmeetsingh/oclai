package cmd

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thejasmeetsingh/oclai/app"
	"github.com/thejasmeetsingh/oclai/mcp"
	"github.com/thejasmeetsingh/oclai/ollama"
	"github.com/thejasmeetsingh/oclai/utils"
)

var rootPath = ""

var (
	rootCmd = &cobra.Command{
		Use:     "oclai",
		Short:   "A completely offline agentic CLI",
		Long:    utils.InfoBox("An offline agentic CLI that brings Claude Code and Gemini CLI capabilities to your terminal using local AI models."),
		Example: `oclai q "What's the latest news of today"`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 0 {
				return
			}

			globalCmds := []string{"baseURL", "model", "ctx"}
			for _, gloglobalCmd := range globalCmds {
				if cmd.Flags().Lookup(gloglobalCmd).Changed {
					return
				}
			}

			cmd.Help()
		},
	}

	modelsCmd = &cobra.Command{
		Use:   "models",
		Short: "List available models",
		Long:  utils.InfoBox("Display all models currently available in your local Ollama installation."),
		Run: func(cmd *cobra.Command, args []string) {
			content, err := ollama.ShowModels(app.OclaiConfig.BaseURL, nil)
			if err != nil {
				fmt.Println(utils.ErrorMessage(fmt.Sprintf("Error listing models: %s", err.Error())))
				os.Exit(1)
			}
			fmt.Println(content)
		},
	}

	statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Check Ollama service status",
		Long:  utils.InfoBox("Check if Ollama service is running and display connection information."),
		Run: func(cmd *cobra.Command, args []string) {
			if err := ollama.CheckOllamaConnection(app.OclaiConfig.BaseURL); err != nil {
				fmt.Println(utils.ErrorMessage(fmt.Sprintf("Ollama Status: %s", err.Error())))
				os.Exit(1)
			}

			fmt.Println(utils.SuccessBox(fmt.Sprintf("Ollama is running at: %s", app.OclaiConfig.BaseURL)))
		},
	}
)

func setBaseURL(arg string) error {
	arg = strings.TrimSpace(arg)

	if arg == "" {
		return fmt.Errorf("✗ 'baseURL' cannot be empty. Please provide a valid URL")
	}

	baseURL, err := url.Parse(strings.TrimSpace(arg))
	if err != nil {
		return fmt.Errorf("✗ invalid URL format: %s. Please enter a valid URL", err.Error())
	}

	app.OclaiConfig.BaseURL = baseURL.String()
	if err = app.UpdateConfig(rootPath); err != nil {
		return err
	}

	fmt.Println(utils.SuccessBox("BaseURL updated successfully!"))

	return nil
}

func setDefaultModel(arg string) error {
	arg = strings.TrimSpace(arg)

	if arg == "" {
		return fmt.Errorf("✗ model value cannot be empty. Please provide a valid model name")
	}

	app.OclaiConfig.DefaultModel = strings.TrimSpace(arg)
	if err := app.UpdateConfig(rootPath); err != nil {
		return err
	}

	fmt.Println(utils.SuccessBox("Default Model updated successfully!"))

	return nil
}

func setNumCtx(arg string) error {
	arg = strings.TrimSpace(arg)

	if arg == "" {
		return fmt.Errorf("✗ 'num_ctx' value should not be empty")
	}

	numCtx, err := strconv.Atoi(arg)
	if err != nil {
		return fmt.Errorf("✗ value should be a valid integer")
	}

	app.OclaiConfig.NumCtx = numCtx
	if err := app.UpdateConfig(rootPath); err != nil {
		return err
	}

	fmt.Println(utils.SuccessBox("NumCtx value updated successfully!"))

	return nil
}

func init() {
	_rootPath, err := utils.GetAppRootDir()
	if err != nil {
		fmt.Println(utils.ErrorMessage(fmt.Sprintf("Error caught while retreiving root path: %s", err.Error())))
		os.Exit(1)
	}

	rootPath = _rootPath

	rootCmd.PersistentFlags().Func("baseURL", "Set Ollama BaseURL", setBaseURL)
	rootCmd.PersistentFlags().Func("model", "Set Default Model", setDefaultModel)
	rootCmd.PersistentFlags().Func("ctx", "Set Context Limit", setNumCtx)

	rootCmd.AddCommand(
		modelsCmd,
		statusCmd,
		app.Query,
		app.Chat,
		mcp.McpRootCmd,
	)
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(utils.ErrorMessage(fmt.Sprintf("Error executing command: %s", err.Error())))
		os.Exit(1)
	}
}
