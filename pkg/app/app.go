package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fatih/color"
	"github.com/thejasmeetsingh/oclai/pkg/config"
	"github.com/thejasmeetsingh/oclai/pkg/markdown"
)

type (
	role string

	ToolCall struct {
		Function struct {
			Name string         `json:"name"`
			Args map[string]any `json:"arguments"`
		} `json:"function"`
	}

	ChatMessage struct {
		Role      role       `json:"role"`
		Content   string     `json:"content,omitempty"`
		ToolName  string     `json:"tool_name,omitempty"`
		ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	}

	ModelResponse struct {
		Model              string      `json:"model"`
		CreatedAt          time.Time   `json:"created_at"`
		Done               bool        `json:"done"`
		Message            ChatMessage `json:"message"`
		Context            []int       `json:"context,omitempty"`
		TotalDuration      int64       `json:"total_duration,omitempty"`
		LoadDuration       int64       `json:"load_duration,omitempty"`
		PromptEvalCount    int         `json:"prompt_eval_count,omitempty"`
		PromptEvalDuration int64       `json:"prompt_eval_duration,omitempty"`
		EvalCount          int         `json:"eval_count,omitempty"`
		EvalDuration       int64       `json:"eval_duration,omitempty"`
	}

	ModelInfo struct {
		Name       string    `json:"model"`
		Size       int64     `json:"size"`
		ModifiedAt time.Time `json:"modified_at"`
	}

	ModelsResponse struct {
		Models []ModelInfo `json:"models"`
	}
)

const (
	System    role = "system"
	User      role = "user"
	Assistant role = "assistant"
	Tool      role = "tool"
)

func (r role) isValidRole() bool {
	switch r {
	case System, User, Assistant, Tool:
		return true
	default:
		return false
	}
}

func (r role) toString() string {
	return string(r)
}

func CheckOllamaConnection() error {
	url := config.OclaiConfig.BaseURL
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

func ListModels() ([]ModelInfo, error) {
	resp, err := http.Get(config.OclaiConfig.BaseURL + "/api/tags")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var modelsResp ModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, err
	}

	return modelsResp.Models, nil
}

func Chat(messages []ChatMessage) error {
	body := &bytes.Buffer{}
	encoder := json.NewEncoder(body)
	encoder.Encode(messages)

	response, err := http.Post(config.OclaiConfig.BaseURL+"/api/chat", "application/json", body)
	if err != nil {
		return nil
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status %d", response.StatusCode)
	}

	var modelResponse ModelResponse
	if err = json.NewDecoder(response.Body).Decode(&modelResponse); err != nil {
		return fmt.Errorf("error while parsing the model response: %s", err.Error())
	}

	if err = markdown.Render(modelResponse.Message.Content); err != nil {
		return fmt.Errorf("error caught while rendering response: %s", err.Error())
	}

	if modelResponse.Done {
		if modelResponse.TotalDuration > 0 {
			duration := time.Duration(modelResponse.TotalDuration)
			tokensPerSec := float64(modelResponse.EvalCount) / duration.Seconds()

			color.New(color.FgGreen).Printf("✓ Generated %d tokens in %v (%.1f tokens/sec)\n",
				modelResponse.EvalCount, duration, tokensPerSec)
		}
	}

	return nil
}
