package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	goMCP "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/thejasmeetsingh/oclai/ollama"
)

// listTools retrieves the list of available tools from the MCP client and converts them
// to the ollama.Tool format for compatibility.
// It handles the conversion of tool parameters from JSON format to the ollama.Parameter type.
func listTools(ctx context.Context, cs *goMCP.ClientSession) ([]ollama.Tool, error) {
	var (
		tools  []ollama.Tool
		params ollama.Parameter
	)

	// Fetch the list of tools from the MCP client
	mcpTools, err := cs.ListTools(ctx, nil)
	if err != nil {
		return tools, err
	}

	// Process each tool to convert it to the ollama.Tool format
	for _, tool := range mcpTools.Tools {
		// Marshal the tool's input schema to JSON
		inputSchema, err := tool.InputSchema.MarshalJSON()
		if err != nil {
			return tools, err
		}

		// Unmarshal the JSON into the ollama.Parameter type
		err = json.Unmarshal(inputSchema, &params)
		if err != nil {
			return tools, err
		}

		// Add the converted tool to the list
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

// CallTool executes a specific tool using the MCP client session and returns the results.
// It handles the execution of the tool and processes the results to return them as a string.
func CallTool(ctx context.Context, cs *goMCP.ClientSession, params *goMCP.CallToolParams) (string, error) {
	// Execute the tool with the provided parameters
	result, err := cs.CallTool(ctx, params)
	if err != nil {
		return "", err
	}

	// If the tool execution resulted in an error, return an error message
	if result.IsError {
		return "", fmt.Errorf("tool execution failed")
	}

	// Extract the text content from the tool's result
	var toolResults []string
	for _, content := range result.Content {
		toolResults = append(toolResults, content.(*goMCP.TextContent).Text)
	}

	// Join the results into a single string and return
	return strings.Join(toolResults, "."), nil
}

// GetAllTools returns all available tools from the MCP servers.
// It aggregates tools from all servers to provide a comprehensive list.
func GetAllTools() []ollama.Tool {
	tools := make([]ollama.Tool, 0)
	servers := mcpServers["servers"]

	for _, server := range servers {
		tools = append(tools, server.Tools...)
	}

	return tools
}

// GetSessionFromToolName retrieves a MCP client session for the specified tool name.
func GetSessionFromToolName(ctx context.Context, toolName string) (*goMCP.ClientSession, error) {
	servers := mcpServers["servers"]

	// Search for the tool across all servers
	for _, server := range servers {
		for _, tool := range server.Tools {
			if strings.EqualFold(tool.Function.Name, toolName) {
				// Create a session for the found server
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
