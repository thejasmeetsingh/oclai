// Provides functionality for interacting with the Ollama service,
// including checking connections, listing models, and making chat requests.

package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fatih/color"
)

type (
	Parameter struct {
		ParameterType string         `json:"type"`
		Properties    map[string]any `json:"properties"`
		Required      []string       `json:"required"`
	}

	Function struct {
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Parameter   Parameter `json:"parameters"`
	}

	Tool struct {
		ToolType string   `json:"type"`
		Function Function `json:"function"`
	}

	// ToolCall represents a function call made by the assistant
	ToolCall struct {
		Function struct {
			Name string         `json:"name"`
			Args map[string]any `json:"arguments"`
		} `json:"function"`
	}

	// Message represents a chat message with role and content
	Message struct {
		Role      string     `json:"role"`
		Content   string     `json:"content,omitempty"`
		Thinking  string     `json:"thinking,omitempty"`
		ToolName  string     `json:"tool_name,omitempty"`
		ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	}

	// ModelRequest represents a request to the model API
	ModelRequest struct {
		Model    string     `json:"model"`
		Think    bool       `json:"think"`
		Format   string     `json:"format,omitempty"`
		Messages *[]Message `json:"messages"`
		Tools    []Tool     `json:"tools,omitempty"`
	}

	// ModelResponse represents the response from the model API
	ModelResponse struct {
		Model              string    `json:"model"`
		CreatedAt          time.Time `json:"created_at"`
		Done               bool      `json:"done"`
		Message            Message   `json:"message"`
		TotalDuration      int64     `json:"total_duration,omitempty"`
		LoadDuration       int64     `json:"load_duration,omitempty"`
		PromptEvalCount    int       `json:"prompt_eval_count,omitempty"`
		PromptEvalDuration int64     `json:"prompt_eval_duration,omitempty"`
		EvalCount          int       `json:"eval_count,omitempty"`
		EvalDuration       int64     `json:"eval_duration,omitempty"`
	}

	// ModelInfo contains information about a model
	ModelInfo struct {
		Name       string    `json:"model"`
		Size       int64     `json:"size"`
		ModifiedAt time.Time `json:"modified_at"`
	}

	// ModelsResponse contains a list of model information
	ModelsResponse struct {
		Models []ModelInfo `json:"models"`
	}
)

const (
	SystemRole    string = "system"
	UserRole      string = "user"
	AssistantRole string = "assistant"
	ToolRole      string = "tool"
)

// CheckOllamaConnection checks if the Ollama service is running
func CheckOllamaConnection() error {
	url := OclaiConfig.BaseURL
	resp, err := http.Get(url + "/api/tags")
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama service at %s: %v\nPlease ensure Ollama is running", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama service returned status %d", resp.StatusCode)
	}

	return nil
}

func SystemPromptMessage() Message {
	return Message{
		Role:    SystemRole,
		Content: "You are a helpful Assistant!",
	}
}

// ListModels retrieves a list of available models from the Ollama service
func ListModels() ([]ModelInfo, error) {
	resp, err := http.Get(OclaiConfig.BaseURL + "/api/tags")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama service returned status %d", resp.StatusCode)
	}

	var modelsResp ModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, err
	}

	return modelsResp.Models, nil
}

// Fetch models from ollama and return them in an appropriate format
func ShowModels(models *[]ModelInfo) (string, error) {
	if models == nil {
		newModels, err := ListModels()
		if err != nil {
			return "", err
		}
		models = &newModels
	}

	if len(*models) == 0 {
		return "", fmt.Errorf("no models found. Please install a model using: ollama pull <model-name>")
	}

	var modelMsgs []string

	for idx, model := range *models {
		sizeGB := float64(model.Size) / (1024 * 1024 * 1024)
		modelMsg := fmt.Sprintf("%d. **%s** (%.1f GB) - Modified At: %s", idx+1, model.Name, sizeGB, model.ModifiedAt.Format("2006-01-02 15:04:05"))
		modelMsgs = append(modelMsgs, modelMsg)
	}

	content := fmt.Sprintf("# 📋 Available Models\n%s", strings.Join(modelMsgs, "\n"))
	return RenderMD(content)
}

// Chat sends a chat request to the model API and processes the response
func Chat(request ModelRequest, showStats bool) (string, error) {
	body := &bytes.Buffer{}
	encoder := json.NewEncoder(body)
	encoder.Encode(request)

	response, err := http.Post(OclaiConfig.BaseURL+"/api/chat", "application/json", body)
	if err != nil {
		return "", nil
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d", response.StatusCode)
	}

	var modelResponse ModelResponse
	if err = json.NewDecoder(response.Body).Decode(&modelResponse); err != nil {
		return "", fmt.Errorf("error while parsing the model response: %s", err.Error())
	}

	if modelResponse.Done {
		content := modelResponse.Message.Content
		result, err := RenderMD(content)
		if err != nil {
			return "", err
		}

		if showStats && modelResponse.TotalDuration > 0 {
			duration := time.Duration(modelResponse.TotalDuration)
			tokensPerSec := float64(modelResponse.EvalCount) / duration.Seconds()
			stat := color.New(color.FgGreen).Sprintf("✓ Generated %d tokens in %v (%.1f tokens/sec)\n\n",
				modelResponse.EvalCount, duration, tokensPerSec)

			result = fmt.Sprintf("%s\n%s", result, stat)
		}

		return result, nil
	}

	return "", nil
}
