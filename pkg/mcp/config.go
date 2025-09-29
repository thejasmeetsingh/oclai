package mcp

import (
	"encoding/json"
	"path/filepath"

	goMCP "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/viper"
	"github.com/thejasmeetsingh/oclai/pkg/ollama"
	"github.com/thejasmeetsingh/oclai/pkg/utils"
)

// McpConfigFileName is the name of the configuration file
const McpConfigFileName = "mcp"

// McpServer represents the MCP configuration structure
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
	// A common MCP client instance
	Client = goMCP.NewClient(&goMCP.Implementation{Name: "oclai", Version: "v1.0.0"}, nil)

	// mcpServers map contains all the mcp servers with their respected tool details
	mcpServers = make(map[string][]*McpServer)
)

// LoadConfig initializes and loads the MCP servers configuration
func LoadConfig(rootPath string) error {
	// Construct the full path to the configuration file
	filePath := filepath.Join(rootPath, McpConfigFileName)

	// Retrieve default servers
	servers := getDefaultServers()

	// Initialize a new Viper instance
	v := viper.New()

	// Set the name of the MCP configuration file (without extension)
	v.SetConfigName(McpConfigFileName)
	v.SetConfigType("json")

	// Add the root path as a MCP configuration search path
	v.AddConfigPath(rootPath)

	v.SetDefault("servers", servers)

	// Write the configuration to the file (safe write to avoid overwriting)
	v.SafeWriteConfigAs(filePath)
	if err := v.ReadInConfig(); err != nil {
		return err
	}

	// Read the configuration file content
	data, err := utils.ReadConfig(filePath)
	if err != nil {
		return err
	}

	// Unmarshal the JSON data into the OclaiConfig struct
	return json.Unmarshal(data, &mcpServers)
}

// UpdateConfig updates the MCP configuration file with the current server details
func UpdateConfig(rootPath string) error {
	// Construct the full path to the configuration file
	filePath := filepath.Join(rootPath, McpConfigFileName)

	// Marshal the MCP configuration into JSON format
	data, err := json.MarshalIndent(&mcpServers, "", "  ")
	if err != nil {
		return err
	}

	// Write the JSON data to the configuration file
	return utils.WriteFileContents(filePath, data)
}
