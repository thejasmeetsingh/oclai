package chat

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/thejasmeetsingh/oclai/pkg/app"
)

var (
	infoMSg    = color.New(color.FgBlue)
	successMsg = color.New(color.FgGreen)
	errMsg     = color.New(color.FgRed)
	warningMsg = color.New(color.FgYellow)
)

func showInteractiveHelp() {
	infoMSg.Println("\n📖 Interactive Mode Commands:")
	fmt.Println("  exit, quit     - Exit the interactive session")
	fmt.Println("  /help          - Show this help message")
	fmt.Println("  /models        - List available models")
	fmt.Println("  /model <name>  - Switch to a different model")
	fmt.Println()
}

func startInteractiveSession(model string) error {
	successMsg.Println("🚀 Starting interactive session with", model)
	warningMsg.Println("Type 'exit', 'quit', or press Ctrl+C to end the session")
	warningMsg.Println("Type '/help' for available commands")

	scanner := bufio.NewScanner(os.Stdin)

	models, err := app.ListModels()
	if err != nil {
		return err
	}

	modelRequest := app.ModelRequest{
		Model: model,
		Think: false,
	}

	messages := &[]app.Message{{
		Role:    app.System,
		Content: "You are helpful assistant",
	}}

	for {
		infoMSg.Print("\n💬 You: ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			continue
		}

		// Handle special commands
		switch input {
		case "exit", "quit":
			successMsg.Println("👋 Goodbye!")
			fmt.Println()
			return nil
		case "/help":
			showInteractiveHelp()
			continue
		case "/models":
			models, err := app.ListModels()
			if err != nil {
				return err
			}
			if err := app.ShowModels(&models); err != nil {
				return err
			}
			continue
		}

		// Check for model switch command
		if after, ok := strings.CutPrefix(input, "/model "); ok {
			newModel := strings.TrimSpace(after)
			if newModel != "" {
				var isPresent bool
				for _, _model := range models {
					if newModel == _model.Name {
						isPresent = true
						break
					}
				}

				if !isPresent {
					errMsg.Println("Model does not exists")
					continue
				}

				modelRequest.Model = newModel
				successMsg.Println("✓ Switched to model:", newModel)
			}
			continue
		}

		*messages = append(*messages, app.Message{
			Role:    app.User,
			Content: input,
		})

		// Generate response
		modelRequest.Messages = messages
		if err := app.Chat(modelRequest); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
