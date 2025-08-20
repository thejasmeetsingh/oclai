package chat

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thejasmeetsingh/oclai/pkg/app"
	"github.com/thejasmeetsingh/oclai/pkg/config"
)

var (
	infoMsg = config.InfoMessage
	errMSg  = config.ErrorMessage
)

var Chat = &cobra.Command{
	Use:     "chat",
	Aliases: []string{"ch"},
	Short:   "Start an interactive chat session",
	Long:    infoMsg.Sprint("Start an interactive chat session with the specified model. If no model is specified, you'll be prompted to choose one."),
	Example: `
		oclai chat
		oclai ch
		oclai chat --model gemma3:latest
	`,
	Run: func(cmd *cobra.Command, args []string) {
		var model string

		if config.OclaiConfig.DefaultModel != "" {
			model = config.OclaiConfig.DefaultModel
		} else {
			// List available models and let user choose
			models, err := app.ListModels()
			if err != nil {
				errMSg.Println("Error listing models:", err)
				os.Exit(1)
			}

			if err = app.ShowModels(); err != nil {
				errMSg.Println(err)
				os.Exit(1)
			}

			var choice int

			for {
				infoMsg.Printf("Select a model [%d-%d]: ", 1, len(models))
				scanner := bufio.NewScanner(os.Stdin)
				if scanner.Scan() {
					input := strings.TrimSpace(scanner.Text())

					// Try to parse as number first
					choice, err = strconv.Atoi(input)
					if err != nil {
						errMSg.Println("Invalid input. Please try again")
						continue
					}

					if choice < 1 || choice > len(models) {
						errMSg.Println("Invalid choice. Please try again")
						continue
					}
				}
				break
			}

			model = models[choice-1].Name
		}

		if err := startInteractiveSession(model); err != nil {
			errMSg.Println("Error caught in interactive session:", err)
			os.Exit(1)
		}
	},
}
