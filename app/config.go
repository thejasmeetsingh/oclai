// Contains the implementation for managing the application's configuration.
// It defines the Config struct and provides functions to load, update, and check the configuration.
// The configuration is stored in a JSON file located in the user's home directory.

package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/viper"
	"github.com/thejasmeetsingh/oclai/mcp"
)

const configFileName string = ".oclai-config"

// Config struct holds configuration parameters for the application.
type Config struct {
	// BaseURL is the base URL used for API requests.
	BaseURL string `json:"baseURL"`

	// DefaultModel specifies the default model to be used in the application.
	DefaultModel string `json:"defaultModel"`

	// File is the file path or identifier associated with the configuration.
	File string `json:"file"`

	McpServers map[string]map[string]any `json:"mcpServers"`
}

// OclaiConfig is a global variable that holds the application's configuration.
var OclaiConfig Config

// setupConfig initializes the configuration by setting the default values and writing the configuration file.
func setupConfig() error {
	configFilePath := filepath.Join(os.Getenv("HOME"), configFileName+".json")

	viper.SetConfigName(configFileName) // Name of the config file (without extension)
	viper.SetConfigType("json")         // Config file type
	viper.AddConfigPath("$HOME")        // Add the home directory as a config path

	viper.SetDefault("baseURL", "http://localhost:11434") // Set default base URL
	viper.SetDefault("defaultModel", "")                  // Set default model to empty string
	viper.SetDefault("file", configFilePath)              // Set default file path
	viper.SetDefault("mcpServers", mcp.DefaultServers)

	viper.SafeWriteConfigAs(configFilePath) // Write the config file to the specified file path
	return viper.ReadInConfig()             // Read the configuration file
}

// LoadConfig loads the configuration from the file into the OclaiConfig variable.
func LoadConfig() error {
	err := setupConfig()
	if err != nil {
		return err
	}

	file := viper.GetString("file") // Get the file path from the configuration
	data, err := os.ReadFile(file)  // Read the configuration file
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &OclaiConfig) // Unmarshal the JSON data into the OclaiConfig struct
}

// UpdateConfig updates the configuration file with the current OclaiConfig values.
func UpdateConfig() error {
	data, err := json.MarshalIndent(&OclaiConfig, "", "  ") // Marshal the OclaiConfig struct into JSON format with indentation
	if err != nil {
		return err
	}

	return os.WriteFile(OclaiConfig.File, data, 0644) // Write the updated configuration to the file
}

// DefaultModelCheck checks if the default model is set in the configuration.
func DefaultModelCheck() error {
	if OclaiConfig.DefaultModel == "" {
		msg := color.New(color.FgRed).Sprint("please select a default model 🤖")
		return fmt.Errorf("%s", msg) // Return an error if the default model is not set
	}
	return nil
}
