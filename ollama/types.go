package ollama

import "time"

const (
	SystemRole    string = "system"
	UserRole      string = "user"
	AssistantRole string = "assistant"
	ToolRole      string = "tool"
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
		Model    string         `json:"model"`
		Think    bool           `json:"think"`
		Format   string         `json:"format,omitempty"`
		Messages *[]Message     `json:"messages"`
		Tools    []Tool         `json:"tools,omitempty"`
		Options  map[string]any `json:"options,omitempty"`
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
