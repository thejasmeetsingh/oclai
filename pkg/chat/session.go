package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	"github.com/thejasmeetsingh/oclai/pkg/app"
)

type (
	sessionMessage struct {
		sender  app.Role
		content string
	}

	session struct {
		textInput    textinput.Model
		spinner      spinner.Model
		modelRequest app.ModelRequest
		models       []app.ModelInfo
		messages     []sessionMessage
		history      []string
		waiting      bool
		historyIdx   int
	}

	commandInfo struct {
		name        string
		description string
	}
)

var (
	infoMsg    = color.New(color.FgBlue)
	successMsg = color.New(color.FgGreen)
	errMsg     = color.New(color.FgRed)
	warningMsg = color.New(color.FgYellow)

	subcommands = map[string]commandInfo{
		"/help": {
			name:        "/help",
			description: "Show help message",
		},
		"/clear": {
			name:        "/clear",
			description: "Clear the chat history",
		},
		"/models": {
			name:        "/models",
			description: "List available models",
		},
		"/model": {
			name:        "/model <name>",
			description: "Switch to a different model",
		},
	}
)

func userPromptText() string {
	return color.New(color.FgCyan, color.Bold).Sprint("💬 You: ")
}

func initSession(modelRequest app.ModelRequest, models []app.ModelInfo) *session {
	ti := textinput.New()
	ti.Placeholder = "Type your message here... (try typing '/' for commands)"
	ti.Prompt = userPromptText()
	ti.Focus()
	ti.Width = 80
	ti.ShowSuggestions = true

	spinr := spinner.New()
	spinr.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	spinr.Spinner = spinner.Pulse

	return &session{
		textInput:    ti,
		spinner:      spinr,
		models:       models,
		modelRequest: modelRequest,
		messages:     make([]sessionMessage, 0),
		waiting:      false,
		history:      make([]string, 0),
		historyIdx:   0,
	}
}

func (s *session) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, s.spinner.Tick)
}

func (s *session) clearInput() {
	s.textInput.SetValue("")
	s.textInput.SetSuggestions([]string{})
}

func (s *session) addModelMessage(message app.Message) {
	*s.modelRequest.Messages = append(*s.modelRequest.Messages, message)
}

func (s *session) addSessionMessage(message sessionMessage) {
	s.messages = append(s.messages, message)
}

func handleHelp(s *session) (*session, tea.Cmd) {
	helpText := "📚 Available Commands:\n\n"
	for _, cmd := range subcommands {
		helpText += fmt.Sprintf(" %s - %s\n", cmd.name, cmd.description)
	}
	helpText += "\n💡 Tips:\n"
	helpText += "  • Type / to see available commands with autocomplete\n"
	helpText += "  • Use Ctrl+N/Ctrl+P to navigate suggestions\n"
	helpText += "  • Use ↑↓ arrows to access message history\n"
	helpText += "  • Use ←→ arrows and all standard text editing keys\n"
	helpText += "  • Type 'exit' or 'quit' to leave"

	s.addSessionMessage(sessionMessage{
		sender:  app.System,
		content: infoMsg.Sprint(helpText),
	})
	s.clearInput()

	return s, nil
}

func handleClearHistory(s *session) (*session, tea.Cmd) {
	s.modelRequest.Messages = &[]app.Message{app.SystemPromptMessage()}
	s.messages = make([]sessionMessage, 0)

	s.addSessionMessage(sessionMessage{
		sender:  app.System,
		content: successMsg.Sprint("Chat History cleared! 🧹"),
	})
	s.clearInput()

	return s, nil
}

func handleModelListing(s *session) (*session, tea.Cmd) {
	modelsContent, err := app.ShowModels(&s.models)
	if err != nil {
		s.addSessionMessage(sessionMessage{
			sender:  app.System,
			content: errMsg.Sprint(err.Error()),
		})
		return s, nil
	}

	s.addSessionMessage(sessionMessage{
		sender:  app.System,
		content: modelsContent,
	})
	s.clearInput()

	return s, nil
}

func handleModelSwitch(s *session, newModel string) (*session, tea.Cmd) {
	var isValid bool

	for _, _model := range s.models {
		if newModel == _model.Name {
			isValid = true
			break
		}
	}

	if !isValid {
		s.addSessionMessage(sessionMessage{
			sender:  app.System,
			content: errMsg.Sprint("Model does not exists"),
		})
		return s, nil
	}

	s.modelRequest.Model = newModel
	s.addSessionMessage(sessionMessage{
		sender:  app.System,
		content: successMsg.Sprint("✓ Switched to model:", newModel),
	})
	s.clearInput()

	return s, nil
}

func (s *session) addExitMsg() {
	s.addSessionMessage(sessionMessage{
		sender:  app.System,
		content: successMsg.Sprint("👋 Goodbye!"),
	})
	s.clearInput()
}

func (s *session) updateSuggestions() {
	input := s.textInput.Value()

	if !strings.HasPrefix(input, "/") || len(input) <= 1 {
		s.textInput.SetSuggestions([]string{})
		return
	}

	var matches []string
	for cmd := range subcommands {
		if strings.HasPrefix(cmd, input) {
			matches = append(matches, cmd)
		}
	}

	s.textInput.SetSuggestions(matches)
}

func (s *session) handleCommand(command string) (*session, tea.Cmd) {
	cmd := strings.Fields(command)

	if cmdInfo, exists := subcommands[cmd[0]]; exists {
		switch cmdInfo.name {
		case "/help":
			return handleHelp(s)
		case "/clear":
			return handleClearHistory(s)
		case "/models":
			return handleModelListing(s)
		case "/model":
			if len(cmd) != 2 {
				break
			}
			return handleModelSwitch(s, cmd[1])
		}
	}

	s.addSessionMessage(sessionMessage{
		sender:  app.System,
		content: errMsg.Sprintf("Unknown command: %s. Type '/help' for available commands.", cmd),
	})
	s.clearInput()

	return s, nil
}

func (s *session) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			s.addExitMsg()
			return s, tea.Quit

		case "enter":
			if s.waiting {
				return s, s.spinner.Tick
			}

			input := strings.ToLower(strings.TrimSpace(s.textInput.Value()))
			if input == "" {
				return s, nil
			}

			if input == "exit" || input == "quit" {
				s.addExitMsg()
				return s, tea.Quit
			}

			s.history = append(s.history, input)
			s.historyIdx = len(s.history)

			if strings.HasPrefix(input, "/") {
				return s.handleCommand(input)
			}

			s.addModelMessage(app.Message{
				Role:    app.User,
				Content: input,
			})
			s.addSessionMessage(sessionMessage{
				sender:  app.User,
				content: input,
			})

			s.waiting = true
			s.clearInput()

			response, err := app.Chat(s.modelRequest, false)
			if err != nil {
				s.addSessionMessage(sessionMessage{
					sender:  app.System,
					content: errMsg.Sprint(err.Error()),
				})
				return s, nil
			}

			s.addModelMessage(app.Message{
				Role:    app.Assistant,
				Content: response,
			})
			s.addSessionMessage(sessionMessage{
				sender:  app.Assistant,
				content: response,
			})
			s.waiting = false

			return s, nil

		case "up":
			if len(s.history) > 0 && s.historyIdx > 0 {
				s.historyIdx--
				s.textInput.SetValue(s.history[s.historyIdx])
			}
			return s, nil

		case "down":
			if len(s.history) > 0 && s.historyIdx < len(s.history)-1 {
				s.historyIdx++
				s.textInput.SetValue(s.history[s.historyIdx])
			} else if s.historyIdx == len(s.history)-1 {
				s.historyIdx = len(s.history)
				s.textInput.SetValue("")
			}
			return s, nil
		}

	case spinner.TickMsg:
		s.spinner, cmd = s.spinner.Update(msg)
		return s, cmd
	}

	prevValue := s.textInput.Value()
	s.textInput, cmd = s.textInput.Update(msg)

	if s.textInput.Value() != prevValue {
		s.updateSuggestions()
	}

	return s, cmd
}

func (s *session) View() string {
	var output strings.Builder

	output.WriteString(successMsg.Sprintf("🚀 Starting interactive session with %s\n\n", s.modelRequest.Model))
	output.WriteString(warningMsg.Sprintln("Type 'exit', 'quit', or press Ctrl+C to end the session"))
	output.WriteString(warningMsg.Sprint("Type '/help' for available commands\n\n"))
	output.WriteString(strings.Repeat("-", 70) + "\n\n")

	if len(s.history) == 0 {
		output.WriteString(infoMsg.Sprint("Start the conversation by typing a message below!\n\n"))
	} else {
		for idx := 0; idx < len(s.messages); idx++ {
			msg := s.messages[idx]

			switch msg.sender {
			case app.User:
				output.WriteString(fmt.Sprintf("%s %s\n\n", userPromptText(), msg.content))
			case app.Assistant:
				model := successMsg.Sprint(s.modelRequest.Model)
				output.WriteString(fmt.Sprintf("🤖 %s:\n%s", model, msg.content))
			case app.System:
				output.WriteString(msg.content + "\n\n")
			}
		}
		output.WriteString("\n")
	}

	output.WriteString(strings.Repeat("-", 70) + "\n\n")

	if s.waiting {
		output.WriteString(s.spinner.View())
	} else {
		output.WriteString(s.textInput.View())
	}

	return output.String()
}
