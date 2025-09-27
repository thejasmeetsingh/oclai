package app

import (
	"context"

	goMCP "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/thejasmeetsingh/oclai/mcp"
	"github.com/thejasmeetsingh/oclai/ollama"
)

func getToolResp(ctx context.Context, tool ollama.ToolCall) (string, error) {
	mcpSession, err := mcp.GetSessionFromToolName(ctx, tool.Function.Name)
	if err != nil {
		return "", err
	}

	toolParams := &goMCP.CallToolParams{
		Name:      tool.Function.Name,
		Arguments: tool.Function.Args,
	}

	return mcp.CallTool(ctx, mcpSession, toolParams)
}

func chatWithTools(ctx context.Context, request ollama.ModelRequest) (*ollama.ModelResponse, error) {
	request.Options = map[string]any{"num_ctx": OclaiConfig.NumCtx}

	response, err := ollama.Chat(OclaiConfig.BaseURL, request)
	if err != nil {
		return nil, err
	}

	toolCalls := response.Message.ToolCalls

	if len(toolCalls) != 0 {
		for _, tool := range toolCalls {
			toolResp, err := getToolResp(ctx, tool)
			if err != nil {
				return nil, err
			}

			*request.Messages = append(*request.Messages, ollama.Message{
				Role:    ollama.ToolRole,
				Content: toolResp,
			})
		}

		return chatWithTools(ctx, request)
	}

	return response, nil
}
