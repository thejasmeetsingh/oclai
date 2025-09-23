package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/thejasmeetsingh/oclai/mcp"
	"github.com/thejasmeetsingh/oclai/ollama"
	"github.com/thejasmeetsingh/oclai/utils"
)

var (
	fileContents []string
	mcpTools     []ollama.Tool
)

var (
	Chat = &cobra.Command{
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
				Tools:    mcpTools,
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

	Query = &cobra.Command{
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
			if OclaiConfig.DefaultModel == "" {
				msg := color.New(color.FgRed).Sprint("please select a default model 🤖")
				return fmt.Errorf("%s", msg)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			query := strings.TrimSpace(strings.Join(args, " "))
			errMsg := color.New(color.FgRed)

			if query == "" {
				errMsg.Println("Please provide a query 😒")
				return
			}

			if len(fileContents) == 0 {
				contents, err := utils.ReadPipedInput()
				if err != nil {
					errMsg.Println(err.Error())
					os.Exit(1)
				}

				fileContents = contents
			}

			if len(fileContents) != 0 {
				query = fmt.Sprintf("```\n%s\n```\nUser Query: %s", strings.Join(fileContents, "\n"), query)
			}

			request := ollama.ModelRequest{
				Model: OclaiConfig.DefaultModel,
				Think: false,
				Messages: &[]ollama.Message{{
					Role:    ollama.UserRole,
					Content: query,
				}},
				Tools: mcpTools,
			}

			modelResponse, err := chatWithTools(context.Background(), request, nil)
			if err != nil {
				errMsg.Println(err)
				os.Exit(1)
			}

			result, err := utils.ToMarkDown(modelResponse.Message.Content)
			if err != nil {
				errMsg.Println(err)
				os.Exit(1)
			}

			if modelResponse.TotalDuration > 0 {
				duration := time.Duration(modelResponse.TotalDuration)
				tokensPerSec := float64(modelResponse.EvalCount) / duration.Seconds()
				stat := color.New(color.FgGreen).Sprintf("✓ Generated %d tokens in %v (%.1f tokens/sec)\n\n",
					modelResponse.EvalCount, duration, tokensPerSec)

				result = fmt.Sprintf("%s\n%s", result, stat)
			}

			fmt.Println(result)
		},
	}
)

func init() {
	mcpTools = mcp.GetAllTools()

	Query.PersistentFlags().FuncP("file", "f", "Read from file and ask query about the content", func(s string) error {
		contents, err := utils.ReadFileContent(s)
		if err != nil {
			return err
		}

		fileContents = contents
		return nil
	})
}
