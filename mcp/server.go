package mcp

import (
	"context"
	"fmt"
	"strings"
)

// getDefaultServers returns the default list of servers configured for MCP
func getDefaultServers() []McpServer {
	return []McpServer{
		{
			IsSSE:   false,
			Name:    "filesystem",
			Command: "npx",
			Args: []string{
				"-y",
				"@modelcontextprotocol/server-filesystem",
				".",
			},
		},
		{
			IsSSE:   false,
			Name:    "sequentialthinking",
			Command: "npx",
			Args: []string{
				"-y",
				"@modelcontextprotocol/server-sequential-thinking",
			},
		},
		{
			IsSSE:   false,
			Name:    "fetch",
			Command: "docker",
			Args: []string{
				"run",
				"-i",
				"--rm",
				"mcp/fetch",
			},
		},
	}
}

// InitializeServers sets up the MCP servers with the given context and root path
func InitializeServers(ctx context.Context, rootPath string) error {
	// Retrieve the servers configuration from the mcpServers map
	servers := mcpServers["servers"]

	for _, server := range servers {
		// Check if both Command and Endpoint are provided (at least one is required)
		if server.Command == "" && server.Endpoint == "" {
			return fmt.Errorf("no transport is provided for '%s' server", server.Name)
		}

		// Ensure Args is initialized if it's empty
		if len(server.Args) == 0 {
			server.Args = make([]string, 0)
		}

		// Ensure Headers is initialized if it's empty
		if len(server.Headers) == 0 {
			server.Headers = make(map[string]string)
		}

		// Ensure Env is initialized if it's empty
		if len(server.Env) == 0 {
			server.Env = make(map[string]string)
		}

		// Create a session for the server
		session, err := createSession(ctx, *server)
		if err != nil {
			return err
		}

		// List the available tools for the server
		tools, err := listTools(ctx, session)
		if err != nil {
			return err
		}

		// If tools are available, add them to the server configuration
		if len(tools) != 0 {
			server.Tools = tools
		}

		session.Close()
	}

	// Update the configuration with the current settings
	err := UpdateConfig(rootPath)
	if err != nil {
		return err
	}

	return nil
}

// isServerExists checks if a server with the given name already exists
func isServerExists(newServerName string) int {
	for idx, server := range mcpServers["servers"] {
		if strings.EqualFold(server.Name, newServerName) {
			return idx
		}
	}

	return -1
}

// addServer adds a new server to the configuration and initializes it
func addServer(rootPath string, mcpServer McpServer) error {
	// Check if a server with the same name already exists
	result := isServerExists(mcpServer.Name)
	if result != -1 {
		return fmt.Errorf("server with '%s' name already exists", mcpServer.Name)
	}

	// Add the new server to the servers list
	mcpServers["servers"] = append(mcpServers["servers"], &mcpServer)
	return InitializeServers(context.Background(), rootPath)
}

// removeServer removes a server from the configuration
func removeServer(rootPath, serverName string) error {
	// Find the index of the server with the given name
	idx := isServerExists(serverName)
	if idx == -1 {
		return fmt.Errorf("server with '%s' name does not exists", serverName)
	}

	// Remove the server from the servers list
	mcpServers["servers"] = append(mcpServers["servers"][:idx], mcpServers["servers"][idx+1:]...)
	return InitializeServers(context.Background(), rootPath)
}

// getServerList returns a list of server names
func getServerList() []string {
	servers := make([]string, 0)

	// Iterate over each server and collect their names
	for _, server := range mcpServers["servers"] {
		servers = append(servers, server.Name)
	}

	return servers
}
