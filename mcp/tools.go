package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	goMCP "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/thejasmeetsingh/oclai/ollama"
)

func ListTools(ctx context.Context, cs *goMCP.ClientSession) ([]ollama.Tool, error) {
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

func SaveTools(tools map[string][]ollama.Tool) error {
	toolsFilePath := filepath.Join(os.Getenv("HOME"), ToolsFileName)

	data, err := json.MarshalIndent(tools, "", "\t")
	if err != nil {
		return err
	}

	err = os.WriteFile(toolsFilePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
