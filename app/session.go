package app

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/thejasmeetsingh/oclai/ollama"
	"github.com/thejasmeetsingh/oclai/utils"
)

const (
	height int = 20
	width  int = 100

	infoMsg    messageType = "info"
	successMsg messageType = "success"
	errMsg     messageType = "error"
	usrMsg     messageType = "user"
	aiMsg      messageType = "ai"
)

type (
	messageType string

	sessionMessage struct {
		_type   messageType
		content string
	}

	session struct {
		textInput        textinput.Model
		spinner          spinner.Model
		vp               viewport.Model
		modelRequest     ollama.ModelRequest
		spinnerMsg       string
		messagesMarkdown string
		models           []ollama.ModelInfo
		waiting          bool
	}

	commandInfo struct {
		name        string
		description string
	}
)

var subcommands = map[string]commandInfo{
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

func userPromptText() string {
	return utils.OtherMessage("💬 You: ")
}

func getMarkdownString(content string) string {
	contentMD, err := utils.ToMarkDown(content)
	if err != nil {
		fmt.Println(utils.ErrorMessage(err.Error()))
		os.Exit(1)
	}

	return contentMD
}

func initSession(modelRequest ollama.ModelRequest, models []ollama.ModelInfo) *session {
	ti := textinput.New()
	ti.Placeholder = "Type your message here... (try typing '/' for commands)"
	ti.Prompt = userPromptText()
	ti.Focus()
	ti.Width = width
	ti.ShowSuggestions = true

	spinr := spinner.New()
	spinr.Style = utils.LoaderStyle
	spinr.Spinner = spinner.Points

	vp := viewport.New(width, height)

	return &session{
		textInput:        ti,
		spinner:          spinr,
		vp:               vp,
		models:           models,
		modelRequest:     modelRequest,
		messagesMarkdown: "",
		spinnerMsg:       "",
		waiting:          false,
	}
}

func (s *session) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, s.spinner.Tick)
}

func (s *session) clearInput() {
	s.textInput.Reset()
	s.textInput.SetSuggestions([]string{})
}

func (s *session) addModelMessage(message ollama.Message) {
	*s.modelRequest.Messages = append(*s.modelRequest.Messages, message)
}

func (s *session) updateSessionMessages(message sessionMessage) {
	switch message._type {
	case successMsg:
		message.content = utils.SuccessBox(message.content)
	case errMsg:
		message.content = utils.ErrorBox(message.content)
	case usrMsg:
		message.content = utils.UserMsgBox(message.content)
	}

	s.messagesMarkdown += message.content

	s.vp.SetContent(s.messagesMarkdown)
	s.vp.GotoBottom()
}

func handleHelp(s *session) (*session, tea.Cmd) {
	helpText := "# 📚 Available Commands:\n\n"
	for _, cmd := range subcommands {
		helpText += fmt.Sprintf("- %s : %s\n", cmd.name, cmd.description)
	}
	helpText += "\n## 💡 Tips:\n"
	helpText += "- Type / to see available commands with autocomplete\n"
	helpText += "- Type 'exit' or 'quit' to leave"

	s.updateSessionMessages(sessionMessage{
		_type:   infoMsg,
		content: getMarkdownString(helpText),
	})
	s.clearInput()

	return s, nil
}

func handleClearHistory(s *session) (*session, tea.Cmd) {
	s.modelRequest.Messages = &[]ollama.Message{ollama.SystemPromptMessage()}
	s.messagesMarkdown = ""

	s.updateSessionMessages(sessionMessage{
		_type:   successMsg,
		content: "Chat History cleared! 🧹",
	})
	s.clearInput()

	return s, nil
}

func handleModelListing(s *session) (*session, tea.Cmd) {
	modelsContent, err := ollama.ShowModels(OclaiConfig.BaseURL, &s.models)
	if err != nil {
		s.updateSessionMessages(sessionMessage{
			_type:   errMsg,
			content: err.Error(),
		})
		return s, nil
	}

	s.updateSessionMessages(sessionMessage{
		_type:   infoMsg,
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
		s.updateSessionMessages(sessionMessage{
			_type:   errMsg,
			content: "Model does not exists",
		})
	} else {
		s.modelRequest.Model = newModel
		s.updateSessionMessages(sessionMessage{
			_type:   successMsg,
			content: "Switched to model: " + newModel,
		})
	}

	s.clearInput()
	return s, nil
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

	s.updateSessionMessages(sessionMessage{
		_type:   errMsg,
		content: fmt.Sprintf("Unknown command: %s. Type '/help' to view the available commands.", cmd),
	})
	s.clearInput()

	return s, nil
}

func (s *session) sendChatRequest() {
	modelResponse, err := chatWithTools(context.Background(), s.modelRequest)
	if err != nil {
		s.updateSessionMessages(sessionMessage{
			_type:   errMsg,
			content: err.Error(),
		})
		return
	}

	content, err := utils.ToMarkDown(modelResponse.Message.Content)
	if err != nil {
		s.updateSessionMessages(sessionMessage{
			_type:   errMsg,
			content: err.Error(),
		})
		return
	}

	s.addModelMessage(ollama.Message{
		Role:    ollama.AssistantRole,
		Content: content,
	})
	s.updateSessionMessages(sessionMessage{
		_type:   aiMsg,
		content: content,
	})

	s.waiting = false
	s.spinnerMsg = ""
}

func (s *session) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return s, tea.Quit

		case "down":
			s.vp.ScrollDown(1)
			return s, nil

		case "up":
			s.vp.ScrollUp(1)
			return s, nil

		case "enter":
			if s.waiting {
				return s, s.spinner.Tick
			}

			input := strings.ToLower(strings.TrimSpace(s.textInput.Value()))
			if input == "" {
				return s, nil
			}

			if input == "exit" || input == "quit" {
				return s, tea.Quit
			}

			if strings.HasPrefix(input, "/") {
				return s.handleCommand(input)
			}

			s.addModelMessage(ollama.Message{
				Role:    ollama.UserRole,
				Content: input,
			})
			s.updateSessionMessages(sessionMessage{
				_type:   usrMsg,
				content: input,
			})

			s.waiting = true
			s.spinnerMsg = "Thinking"

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
	cmds = append(cmds, cmd)

	if s.textInput.Value() != prevValue {
		s.updateSuggestions()
	}

	s.vp, cmd = s.vp.Update(msg)
	cmds = append(cmds, cmd)

	return s, tea.Batch(cmds...)
}

func (s *session) View() string {
	var (
		output strings.Builder

		// Sections
		top    string
		middle string
		bottom string
	)

	startupTxt := fmt.Sprintf("# 🚀 Starting interactive session with *%s*\n", s.modelRequest.Model)
	startupTxt += "- Type `exit`, `quit`, or press `Ctrl+C` to end the session.\n"
	startupTxt += "- Type `/help` for available commands."

	top = getMarkdownString(startupTxt)

	if s.messagesMarkdown == "" {
		middle = getMarkdownString("*Start the conversation by typing a message below!*")
	} else {
		middle = s.vp.View()
	}

	if s.waiting {
		bottom = s.spinnerMsg + " " + s.spinner.View()
	} else {
		bottom = s.textInput.View()
	}

	output.WriteString(top + "\n\n" + middle + "\n\n" + bottom)

	return output.String()
}
