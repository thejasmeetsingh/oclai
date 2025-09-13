package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	goMCP "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/thejasmeetsingh/oclai/ollama"
)

func ListTools(ctx context.Context, cs *goMCP.ClientSession) ([]*ollama.Tool, error) {
	mcpTools, err := cs.ListTools(ctx, nil)
	if err != nil {
		return nil, err
	}

	var tools []*ollama.Tool
	var params ollama.Parameter

	for _, tool := range mcpTools.Tools {
		inputSchema, err := tool.InputSchema.MarshalJSON()
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(inputSchema, &params)
		if err != nil {
			return nil, err
		}

		tools = append(tools, &ollama.Tool{
			ToolType: "function",
			Function: ollama.Function{
				Name:        tool.Name,
				Description: tool.Description,
				Parameter:   params,
			},
		})
	}

	return tools, nil
}

func CallTool(ctx context.Context, cs *goMCP.ClientSession, params *goMCP.CallToolParams) (string, error) {
	result, err := cs.CallTool(ctx, params)
	if err != nil {
		return "", err
	}

	if result.IsError {
		return "", fmt.Errorf("tool execution failed")
	}

	var toolResults []string

	for _, content := range result.Content {
		toolResults = append(toolResults, content.(*goMCP.TextContent).Text)
	}

	return strings.Join(toolResults, "."), nil
}
