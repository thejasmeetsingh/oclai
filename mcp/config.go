package mcp

import (
	"fmt"
	"os"
	"path/filepath"

	goMCP "github.com/modelcontextprotocol/go-sdk/mcp"
)

const ToolsFileName = ".oclai-tools.json"

var (
	Client = goMCP.NewClient(&goMCP.Implementation{Name: "oclai", Version: "v1.0.0"}, nil)

	DefaultServers = map[string]map[string]any{
		"filesystem": {
			"command": "docker",
			"args": []string{
				"run",
				"-i",
				"--rm",
				"-v",
				".:/root",
				"mcp/filesystem",
				"/root",
			},
		},
		"memory": {
			"command": "docker",
			"args": []string{
				"run",
				"-i",
				"--rm",
				"-v",
				fmt.Sprintf("%s:/app/dist", filepath.Join(os.Getenv("HOME"), "memory.json")),
			},
		},
		"sequentialthinking": {
			"command": "docker",
			"args": []string{
				"run",
				"--rm",
				"-i",
				"mcp/sequentialthinking",
			},
		},
		"fetch": {
			"command": "docker",
			"args": []string{
				"run",
				"-i",
				"--rm",
				"mcp/fetch",
			},
		},
	}
)
