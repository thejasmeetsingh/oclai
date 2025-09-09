package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	goMCP "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/thejasmeetsingh/oclai/app"
)

func ListTools(ctx context.Context, cs *goMCP.ClientSession) ([]*app.Tool, error) {
	mcpTools, err := cs.ListTools(ctx, nil)
	if err != nil {
		return nil, err
	}

	var tools []*app.Tool
	var params app.Parameter

	for _, tool := range mcpTools.Tools {
		inputSchema, err := tool.InputSchema.MarshalJSON()
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(inputSchema, &params)
		if err != nil {
			return nil, err
		}

		tools = append(tools, &app.Tool{
			ToolType: "function",
			Function: app.Function{
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
