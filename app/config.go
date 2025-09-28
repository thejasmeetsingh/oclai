package app

import (
	"encoding/json"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/thejasmeetsingh/oclai/utils"
)

// AppConfigFileName is the name of the configuration file
const AppConfigFileName = "config"

// Config represents the application configuration structure
type Config struct {
	BaseURL      string `json:"baseURL"`      // Base URL for API endpoints
	DefaultModel string `json:"defaultModel"` // Default model to use
	NumCtx       int    `json:"numCtx"`       // Maximum context length
	InitMCP      bool   `json:"initMCP"`      // Whether to initialize MCP
}

// OclaiConfig holds the loaded configuration for the application
var OclaiConfig Config

// LoadConfig initializes and loads the application configuration
func LoadConfig(rootPath string) error {
	// Construct the full path to the configuration file
	filePath := filepath.Join(rootPath, AppConfigFileName)

	// Initialize a new Viper instance
	v := viper.New()

	// Set the name of the configuration file (without extension)
	v.SetConfigName(AppConfigFileName)
	v.SetConfigType("json")

	// Add the root path as a configuration search path
	v.AddConfigPath(rootPath)

	v.SetDefault("baseURL", "http://localhost:11434")
	v.SetDefault("defaultModel", "")
	v.SetDefault("numCtx", 8000)
	v.SetDefault("initMCP", true)

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
	return json.Unmarshal(data, &OclaiConfig)
}

// UpdateConfig updates the application configuration file with the current settings
func UpdateConfig(rootPath string) error {
	// Construct the full path to the configuration file
	filePath := filepath.Join(rootPath, AppConfigFileName)

	// Marshal the current configuration into JSON format
	data, err := json.MarshalIndent(&OclaiConfig, "", "  ")
	if err != nil {
		return err
	}

	// Write the JSON data to the configuration file
	return utils.WriteFileContents(filePath, data)
}
