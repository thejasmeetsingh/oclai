package app

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/thejasmeetsingh/oclai/ollama"
	"github.com/thejasmeetsingh/oclai/utils"
)

// Constants for message types used in the application
const (
	height int = 20  // Height of the viewport in lines
	width  int = 100 // Width of the viewport in characters

	// messageType is used to differentiate between different types of messages
	infoMsg    messageType = "info"
	successMsg messageType = "success"
	errMsg     messageType = "error"
	usrMsg     messageType = "user"
	aiMsg      messageType = "ai"
)

type (
	messageType string

	// sessionMessage represents a message with a type and content
	sessionMessage struct {
		_type   messageType
		content string
	}

	// session represents the application state for the chat interface
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

	// commandInfo represents information about available commands
	commandInfo struct {
		name        string
		description string
	}
)

// subcommands is a map of available commands and their descriptions
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

// userPromptText returns the placeholder text for the user input field
func userPromptText() string {
	return utils.OtherMessage("💬 You: ")
}

// getMarkdownString converts content to markdown format
func getMarkdownString(content string) string {
	contentMD, err := utils.ToMarkDown(content)
	if err != nil {
		fmt.Println(utils.ErrorMessage(err.Error()))
		os.Exit(1)
	}

	return contentMD
}

// initSession initializes a new session with default settings
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

// Init initializes the session and returns the initial commands
func (s *session) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, s.spinner.Tick)
}

// clearInput resets the text input field and suggestions
func (s *session) clearInput() {
	s.textInput.Reset()
	s.textInput.SetSuggestions([]string{})
}

// addModelMessage adds a new message to the model request
func (s *session) addModelMessage(message ollama.Message) {
	*s.modelRequest.Messages = append(*s.modelRequest.Messages, message)
}

// updateSessionMessages updates the chat history with a new message
func (s *session) updateSessionMessages(message sessionMessage) {
	timestamp := time.Now().Format(time.Kitchen)

	// Format the message based on its type
	switch message._type {
	case successMsg:
		message.content = utils.SuccessBox(message.content)
	case errMsg:
		message.content = utils.ErrorBox(message.content)
	case usrMsg:
		message.content = utils.UserMsgBox(timestamp, getMarkdownString(fmt.Sprintf("*%s*", message.content)))
	case aiMsg:
		message.content = utils.AiMsgBox(timestamp, message.content)
	}

	// Update the chat history with the new message
	s.messagesMarkdown += message.content

	// Update the viewport with the new content and scroll to the bottom
	s.vp.SetContent(s.messagesMarkdown)
	s.vp.GotoBottom()
}

// handleHelp displays the available commands in a formatted message
func handleHelp(s *session) (*session, tea.Cmd) {
	helpText := "# 📚 Available Commands:\n\n"
	for _, cmd := range subcommands {
		helpText += fmt.Sprintf("- %s : %s\n", cmd.name, cmd.description)
	}
	helpText += "\n## 💡 Tips:\n"
	helpText += "- Type / to see available commands with autocomplete\n"
	helpText += "- Type 'exit' or 'quit' to leave"

	// Update the chat history with the help message
	s.updateSessionMessages(sessionMessage{
		_type:   infoMsg,
		content: getMarkdownString(helpText),
	})
	s.clearInput()

	return s, nil
}

// handleClearHistory clears the chat history and resets the model request
func handleClearHistory(s *session) (*session, tea.Cmd) {
	s.modelRequest.Messages = &[]ollama.Message{ollama.SystemPromptMessage()}
	s.messagesMarkdown = ""

	// Update the chat history with a success message
	s.updateSessionMessages(sessionMessage{
		_type:   successMsg,
		content: "Chat History cleared! 🧹",
	})
	s.clearInput()

	return s, nil
}

// handleModelListing lists available models
func handleModelListing(s *session) (*session, tea.Cmd) {
	modelsContent, err := ollama.ShowModels(OclaiConfig.BaseURL, &s.models)
	if err != nil {
		// Update the chat history with an error message
		s.updateSessionMessages(sessionMessage{
			_type:   errMsg,
			content: err.Error(),
		})
		return s, nil
	}

	// Update the chat history with the model list
	s.updateSessionMessages(sessionMessage{
		_type:   infoMsg,
		content: modelsContent,
	})
	s.clearInput()

	return s, nil
}

// handleModelSwitch switches to a different model
func handleModelSwitch(s *session, newModel string) (*session, tea.Cmd) {
	// Check if the model exists
	isExists, err := ollama.IsModelExists(OclaiConfig.BaseURL, newModel, &s.models)
	if err != nil {
		s.updateSessionMessages(sessionMessage{
			_type:   errMsg,
			content: err.Error(),
		})
	}

	// Provide feedback based on whether the model is valid
	if !isExists {
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

// updateSuggestions updates the text input suggestions based on the current input
func (s *session) updateSuggestions() {
	input := s.textInput.Value()

	// Only show suggestions if the input starts with a slash
	if !strings.HasPrefix(input, "/") || len(input) <= 1 {
		s.textInput.SetSuggestions([]string{})
		return
	}

	// Find matching commands
	var matches []string
	for cmd := range subcommands {
		if strings.HasPrefix(cmd, input) {
			matches = append(matches, cmd)
		}
	}

	// Set the suggestions
	s.textInput.SetSuggestions(matches)
}

// handleCommand processes a command input
func (s *session) handleCommand(command string) (*session, tea.Cmd) {
	cmd := strings.Fields(command)

	// Check if the command exists in the subcommands map
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

	// Provide feedback for unknown commands
	s.updateSessionMessages(sessionMessage{
		_type:   errMsg,
		content: fmt.Sprintf("Unknown command: %s. Type '/help' to view the available commands.", cmd),
	})
	s.clearInput()

	return s, nil
}

// sendChatRequest sends a chat request to the AI model
func (s *session) sendChatRequest() {
	modelResponse, err := chatWithTools(context.Background(), s.modelRequest)
	if err != nil {
		// Handle errors by displaying an error message
		s.updateSessionMessages(sessionMessage{
			_type:   errMsg,
			content: err.Error(),
		})
		return
	}

	// Convert the response to markdown and update the chat history
	content, err := utils.ToMarkDown(modelResponse.Message.Content)
	if err != nil {
		s.updateSessionMessages(sessionMessage{
			_type:   errMsg,
			content: err.Error(),
		})
		return
	}

	// Add the AI response to the model request and update the chat history
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

// Update handles application state updates based on the received message
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

			// Add user message to the model request and update chat history
			s.addModelMessage(ollama.Message{
				Role:    ollama.UserRole,
				Content: input,
			})
			s.updateSessionMessages(sessionMessage{
				_type:   usrMsg,
				content: input,
			})

			// Set waiting state and start the chat request
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

	// Update text input and viewport
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

// View renders the application interface
func (s *session) View() string {
	var (
		output strings.Builder

		// Sections
		top    string
		middle string
		bottom string
	)

	// Startup message with application information
	startupTxt := fmt.Sprintf("# 🚀 Starting interactive session with *%s*\n", s.modelRequest.Model)
	startupTxt += "- Type `exit`, `quit`, or press `Ctrl+C` to end the session.\n"
	startupTxt += "- Type `/help` for available commands."

	top = getMarkdownString(startupTxt)

	// Display chat history or a prompt if it's empty
	if s.messagesMarkdown == "" {
		middle = getMarkdownString("*Start the conversation by typing a message below!*")
	} else {
		middle = s.vp.View()
	}

	// Display spinner or text input based on waiting state
	if s.waiting {
		bottom = s.spinnerMsg + " " + s.spinner.View()
	} else {
		bottom = s.textInput.View()
	}

	// Combine all sections and return the final output
	output.WriteString(top + "\n\n" + middle + "\n\n" + bottom)

	return output.String()
}
