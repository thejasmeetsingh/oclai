package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	goMCP "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/thejasmeetsingh/oclai/ollama"
)

func listTools(ctx context.Context, cs *goMCP.ClientSession) ([]ollama.Tool, error) {
	var (
		tools  []ollama.Tool
		params ollama.Parameter
	)

	mcpTools, err := cs.ListTools(ctx, nil)
	if err != nil {
		return tools, err
	}

	for _, tool := range mcpTools.Tools {
		inputSchema, err := tool.InputSchema.MarshalJSON()
		if err != nil {
			return tools, err
		}

		err = json.Unmarshal(inputSchema, &params)
		if err != nil {
			return tools, err
		}

		tools = append(tools, ollama.Tool{
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

func GetAllTools() []ollama.Tool {
	tools := make([]ollama.Tool, 0)
	servers := mcpServers["servers"]

	for _, server := range servers {
		tools = append(tools, server.Tools...)
	}

	return tools
}

func GetSessionFromToolName(ctx context.Context, toolName string) (*goMCP.ClientSession, error) {
	servers := mcpServers["servers"]

	for _, server := range servers {
		for _, tool := range server.Tools {
			if strings.EqualFold(tool.Function.Name, toolName) {
				session, err := createSession(ctx, *server)
				if err != nil {
					return nil, err
				}
				return session, nil
			}
		}
	}

	return nil, fmt.Errorf("'%s' tool does not exists", toolName)
}
