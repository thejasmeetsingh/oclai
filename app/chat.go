package app

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/thejasmeetsingh/oclai/ollama"
)

var Chat = &cobra.Command{
	Use:     "chat",
	Aliases: []string{"ch"},
	Short:   "Start an interactive chat session",
	Long:    color.New(color.FgBlue, color.Bold).Sprint("Start an interactive chat session with the specified model. If no model is specified, you'll be prompted to choose one."),
	Example: `
		oclai chat
		oclai ch
		oclai chat --model gemma3:latest
	`,
	Run: func(cmd *cobra.Command, args []string) {
		model := OclaiConfig.DefaultModel
		errMsg := color.New(color.FgRed)

		models, err := ollama.ListModels(OclaiConfig.BaseURL)
		if err != nil {
			errMsg.Println("Error listing models:", err)
			os.Exit(1)
		}

		if model == "" {
			modelsContent, err := ollama.ShowModels(OclaiConfig.BaseURL, &models)
			if err != nil {
				errMsg.Println(err)
				os.Exit(1)
			}

			fmt.Println(modelsContent)

			var choice int

			for {
				color.New(color.FgCyan).Printf("Select a model [%d-%d]: ", 1, len(models))
				scanner := bufio.NewScanner(os.Stdin)
				if scanner.Scan() {
					input := strings.TrimSpace(scanner.Text())

					// Try to parse as number first
					choice, err = strconv.Atoi(input)
					if err != nil {
						errMsg.Println("Invalid input. Please try again")
						continue
					}

					if choice < 1 || choice > len(models) {
						errMsg.Println("Invalid choice. Please try again")
						continue
					}
				}
				break
			}

			model = models[choice-1].Name
		}

		modelRequest := ollama.ModelRequest{
			Model:    model,
			Think:    false,
			Messages: &[]ollama.Message{ollama.SystemPromptMessage()},
		}

		program := tea.NewProgram(
			initSession(modelRequest, models),
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)

		if _, err := program.Run(); err != nil {
			errMsg.Println(err.Error())
			os.Exit(1)
		}

		fmt.Println("👋 Goodbye!")
	},
}
