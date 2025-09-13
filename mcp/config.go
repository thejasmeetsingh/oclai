package mcp

import (
	"fmt"
	"os"
	"path/filepath"

	goMCP "github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	Client = goMCP.NewClient(&goMCP.Implementation{Name: "oclai", Version: "v1.0.0"}, nil)

	DefaultServers = map[string]any{
		"filesystem": map[string]any{
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
		"memory": map[string]any{
			"command": "docker",
			"args": []string{
				"run",
				"-i",
				"--rm",
				"-v",
				fmt.Sprintf("%s:/app/dist", filepath.Join(os.Getenv("HOME"), "memory.json")),
			},
		},
		"sequentialthinking": map[string]any{
			"command": "docker",
			"args": []string{
				"run",
				"--rm",
				"-i",
				"mcp/sequentialthinking",
			},
		},
		"fetch": map[string]any{
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
