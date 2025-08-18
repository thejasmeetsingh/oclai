package app

import "time"

type (
	role string

	BaseResponse struct {
		Model              string    `json:"model"`
		CreatedAt          time.Time `json:"created_at"`
		Done               bool      `json:"done"`
		Context            []int     `json:"context,omitempty"`
		TotalDuration      int64     `json:"total_duration,omitempty"`
		LoadDuration       int64     `json:"load_duration,omitempty"`
		PromptEvalCount    int       `json:"prompt_eval_count,omitempty"`
		PromptEvalDuration int64     `json:"prompt_eval_duration,omitempty"`
		EvalCount          int       `json:"eval_count,omitempty"`
		EvalDuration       int64     `json:"eval_duration,omitempty"`
	}

	GenerateResponse struct {
		BaseResponse
		Response string `json:"response"`
	}

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

	ChatResponse struct {
		BaseResponse
		Message ChatMessage `json:"message"`
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
	system    role = "system"
	user      role = "user"
	assistant role = "assistant"
	tool      role = "tool"
)

func (r role) isValidRole() bool {
	switch r {
	case system, user, assistant, tool:
		return true
	default:
		return false
	}
}

func (r role) toString() string {
	return string(r)
}
