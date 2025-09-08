package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func ListTools(ctx context.Context, cs *mcp.ClientSession) ([]*Tool, error) {
	mcpTools, err := cs.ListTools(ctx, nil)
	if err != nil {
		return nil, err
	}

	var tools []*Tool
	var params Parameter

	for _, tool := range mcpTools.Tools {
		inputSchema, err := tool.InputSchema.MarshalJSON()
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(inputSchema, &params)
		if err != nil {
			return nil, err
		}

		tools = append(tools, &Tool{
			ToolType: "function",
			Function: Function{
				Name:        tool.Name,
				Description: tool.Description,
				Parameter:   params,
			},
		})
	}

	return tools, nil
}

func CallTool(ctx context.Context, cs *mcp.ClientSession, params *mcp.CallToolParams) (string, error) {
	result, err := cs.CallTool(ctx, params)
	if err != nil {
		return "", err
	}

	if result.IsError {
		return "", fmt.Errorf("tool execution failed")
	}

	var toolResults []string

	for _, content := range result.Content {
		toolResults = append(toolResults, content.(*mcp.TextContent).Text)
	}

	return strings.Join(toolResults, "."), nil
}
