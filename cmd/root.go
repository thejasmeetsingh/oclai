package cmd

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thejasmeetsingh/oclai/pkg/config"
	"github.com/thejasmeetsingh/oclai/pkg/query"
)

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
		globalCmds := []string{"baseURL", "model"}
		for _, gloglobalCmd := range globalCmds {
			if cmd.Flags().Lookup(gloglobalCmd).Changed {
				return
			}
		}

		cmd.Help()
	},
}

func setBaseURL(arg string) error {
	if arg == "" {
		return fmt.Errorf("❌ baseURL should not be empty")
	}

	baseURL, err := url.Parse(strings.TrimSpace(arg))
	if err != nil {
		return fmt.Errorf("❌ please enter a valid URL")
	}

	config.OclaiConfig.BaseURL = baseURL.String()
	return config.UpdateConfig()
}

func setDefaultModel(arg string) error {
	if arg == "" {
		return fmt.Errorf("❌ model value should not be empty")
	}

	config.OclaiConfig.DefaultModel = strings.TrimSpace(arg)
	return config.UpdateConfig()
}

func init() {
	rootCmd.PersistentFlags().Func("baseURL", "Set Ollama BaseURL", setBaseURL)
	rootCmd.PersistentFlags().Func("model", "Set Default Model", setDefaultModel)

	rootCmd.AddCommand(query.Query)

	if err := config.LoadConfig(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Execute runs the root command.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
