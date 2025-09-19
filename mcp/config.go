package mcp

import (
	"encoding/json"
	"path/filepath"

	goMCP "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/viper"
	"github.com/thejasmeetsingh/oclai/ollama"
	"github.com/thejasmeetsingh/oclai/utils"
)

const McpConfigFileName = "mcp"

type McpServer struct {
	IsSSE    bool              `json:"isSSE"`
	Name     string            `json:"name"`
	Command  string            `json:"command,omitempty"`
	Args     []string          `json:"args,omitempty"`
	Endpoint string            `json:"endpoint,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
	Env      map[string]string `json:"env,omitempty"`
	Tools    []ollama.Tool     `json:"tools,omitempty"`
}

var (
	Client     = goMCP.NewClient(&goMCP.Implementation{Name: "oclai", Version: "v1.0.0"}, nil)
	McpServers = make(map[string][]*McpServer)
)

func LoadConfig(rootPath string) error {
	filePath := filepath.Join(rootPath, McpConfigFileName)
	servers := getDefaultServers()

	v := viper.New()
	v.SetConfigName(McpConfigFileName)
	v.SetConfigType("json")

	v.AddConfigPath(rootPath)
	v.SetDefault("servers", servers)

	v.SafeWriteConfigAs(filePath)
	if err := v.ReadInConfig(); err != nil {
		return err
	}

	data, err := utils.ReadFileContents(filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &McpServers)
}

func UpdateConfig(rootPath string) error {
	filePath := filepath.Join(rootPath, McpConfigFileName)

	data, err := json.MarshalIndent(&McpServers, "", "  ")
	if err != nil {
		return err
	}

	return utils.WriteFileContents(filePath, data)
}
