package mcp

import (
	"context"
	"fmt"
	"strings"
)

func getDefaultServers() []McpServer {
	return []McpServer{
		{
			IsSSE:   false,
			Name:    "filesystem",
			Command: "docker",
			Args: []string{
				"run",
				"-i",
				"--rm",
				"-v",
				".:/root",
				"mcp/filesystem",
				"/root",
			},
		},
		{
			IsSSE:   false,
			Name:    "sequentialthinking",
			Command: "docker",
			Args: []string{
				"run",
				"--rm",
				"-i",
				"mcp/sequentialthinking",
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

func InitializeServers(ctx context.Context, rootPath string) error {
	servers := mcpServers["servers"]

	for _, server := range servers {
		if server.Command == "" && server.Endpoint == "" {
			return fmt.Errorf("no transport is provided for %s server", server.Name)
		}

		if len(server.Args) == 0 {
			server.Args = make([]string, 0)
		}

		if len(server.Headers) == 0 {
			server.Headers = make(map[string]string)
		}

		if len(server.Env) == 0 {
			server.Env = make(map[string]string)
		}

		session, err := CreateSession(ctx, *server)
		if err != nil {
			return err
		}

		tools, err := ListTools(ctx, session)
		if err != nil {
			return err
		}

		if len(tools) != 0 {
			server.Tools = tools
		}

		session.Close()
	}

	err := UpdateConfig(rootPath)
	if err != nil {
		return err
	}

	return nil
}

func isServerExists(newServerName string) int {
	for idx, server := range mcpServers["servers"] {
		if strings.EqualFold(server.Name, newServerName) {
			return idx
		}
	}

	return -1
}

func AddServer(rootPath string, mcpServer McpServer) error {
	result := isServerExists(mcpServer.Name)
	if result != -1 {
		return fmt.Errorf("server with '%s' name already exists", mcpServer.Name)
	}

	mcpServers["servers"] = append(mcpServers["servers"], &mcpServer)
	return InitializeServers(context.Background(), rootPath)
}

func RemoveServer(rootPath, serverName string) error {
	idx := isServerExists(serverName)
	if idx != -1 {
		return fmt.Errorf("server with '%s' name does not exists", serverName)
	}

	mcpServers["servers"] = append(mcpServers["servers"][:idx], mcpServers["servers"][idx:]...)
	return InitializeServers(context.Background(), rootPath)
}
