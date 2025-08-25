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

const (
	info      messageType = "info"
	success   messageType = "success"
	eror      messageType = "error"
	user      messageType = "user"
	assistant messageType = "assistant"
)

type (
	messageType string

	sessionMessage struct {
		_type   messageType
		content string
	}

	session struct {
		textInput    textinput.Model
		spinner      spinner.Model
		modelRequest app.ModelRequest
		models       []app.ModelInfo
		messages     []sessionMessage
		waiting      bool
	}

	commandInfo struct {
		name        string
		description string
	}
)

var (
	infoMsg    = color.New(color.FgBlue)
	successMsg = color.New(color.FgGreen)
	errMsg     = color.New(color.FgRed, color.Bold)
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
			name:        "/model",
			description: "Switch to a different model. Usage: /model <modelName>",
		},
	}
)

func userPromptText() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Render("💬 You: ")
}

func initSession(modelRequest app.ModelRequest, models []app.ModelInfo) *session {
	ti := textinput.New()
	ti.Placeholder = "Type your message here... (try typing '/' for commands)"
	ti.Prompt = userPromptText()
	ti.Focus()
	ti.Width = 80
	ti.ShowSuggestions = true

	spinr := spinner.New()
	spinr.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	spinr.Spinner = spinner.Points

	return &session{
		textInput:    ti,
		spinner:      spinr,
		models:       models,
		modelRequest: modelRequest,
		messages:     make([]sessionMessage, 0),
		waiting:      false,
	}
}

func (s *session) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, s.spinner.Tick)
}

func (s *session) clearInput() {
	s.textInput.Reset()
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
	helpText += "  • Type 'exit' or 'quit' to leave"

	s.addSessionMessage(sessionMessage{
		_type:   info,
		content: helpText,
	})
	s.clearInput()

	return s, nil
}

func handleClearHistory(s *session) (*session, tea.Cmd) {
	s.modelRequest.Messages = &[]app.Message{app.SystemPromptMessage()}
	s.messages = make([]sessionMessage, 0)

	s.addSessionMessage(sessionMessage{
		_type:   success,
		content: "Chat History cleared! 🧹",
	})
	s.clearInput()

	return s, nil
}

func handleModelListing(s *session) (*session, tea.Cmd) {
	modelsContent, err := app.ShowModels(&s.models)
	if err != nil {
		s.addSessionMessage(sessionMessage{
			_type:   eror,
			content: err.Error(),
		})
		return s, nil
	}

	s.addSessionMessage(sessionMessage{
		_type:   info,
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
			_type:   eror,
			content: "Model does not exists",
		})
		return s, nil
	}

	s.modelRequest.Model = newModel
	s.addSessionMessage(sessionMessage{
		_type:   success,
		content: "✓ Switched to model: " + newModel,
	})
	s.clearInput()

	return s, nil
}

func (s *session) addExitMsg() {
	s.addSessionMessage(sessionMessage{
		_type:   success,
		content: "👋 Goodbye!",
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
		_type:   eror,
		content: fmt.Sprintf("Unknown command: %s. Type '/help' for available commands.", cmd),
	})
	s.clearInput()

	return s, nil
}

func (s *session) sendChatRequest() {
	response, err := app.Chat(s.modelRequest, false)
	if err != nil {
		s.addSessionMessage(sessionMessage{
			_type:   eror,
			content: err.Error(),
		})
		return
	}

	s.addModelMessage(app.Message{
		Role:    app.Assistant,
		Content: response,
	})
	s.addSessionMessage(sessionMessage{
		_type:   assistant,
		content: response,
	})

	s.waiting = false
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

			if strings.HasPrefix(input, "/") {
				return s.handleCommand(input)
			}

			s.addModelMessage(app.Message{
				Role:    app.User,
				Content: input,
			})
			s.addSessionMessage(sessionMessage{
				_type:   user,
				content: input,
			})

			s.waiting = true
			s.clearInput()
			go s.sendChatRequest()

			return s, nil
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
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
	output.WriteString(warningMsg.Sprint("Type 'exit', 'quit', or press Ctrl+C to end the session\n\n"))
	output.WriteString(warningMsg.Sprint("Type '/help' for available commands\n\n"))
	output.WriteString(strings.Repeat("-", 70) + "\n\n")

	if len(s.messages) == 0 {
		output.WriteString(infoMsg.Sprint("Start the conversation by typing a message below!\n\n"))
	} else {
		for idx := 0; idx < len(s.messages); idx++ {
			msg := s.messages[idx]

			switch msg._type {
			case user:
				output.WriteString(fmt.Sprintf("%s%s\n\n", userPromptText(), msg.content))
			case assistant:
				output.WriteString(fmt.Sprintf("🤖 %s\n", s.modelRequest.Model))
				output.WriteString(msg.content)
			case info:
				output.WriteString(infoMsg.Sprint(msg.content) + "\n\n")
			case success:
				output.WriteString(successMsg.Sprint(msg.content) + "\n\n")
			case eror:
				output.WriteString(errMsg.Sprint(msg.content) + "\n\n")
			}
		}
		output.WriteString("\n")
	}

	output.WriteString(strings.Repeat("-", 70) + "\n\n")

	if s.waiting {
		output.WriteString(fmt.Sprintf("Thinking %s", s.spinner.View()))
	} else {
		output.WriteString(s.textInput.View())
	}

	return output.String()
}
