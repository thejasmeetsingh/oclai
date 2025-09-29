package main

import (
	"context"
	"fmt"
	"os"

	"github.com/thejasmeetsingh/oclai/pkg/app"
	"github.com/thejasmeetsingh/oclai/pkg/cmd"
	"github.com/thejasmeetsingh/oclai/pkg/mcp"
	"github.com/thejasmeetsingh/oclai/pkg/utils"
)

func main() {
	// Create a background context for the application
	ctx := context.Background()

	// Retrieve the root directory path for the application
	rootPath, err := utils.GetAppRootDir()
	if err != nil {
		fmt.Println(utils.ErrorMessage(fmt.Sprintf("Error caught while retreiving root path: %s", err.Error())))
		os.Exit(1)
	}

	// Load App configuration file
	if err := app.LoadConfig(rootPath); err != nil {
		fmt.Println(utils.ErrorMessage(fmt.Sprintf("Failed to load configuration: %s", err.Error())))
		os.Exit(1)
	}

	// Load and Initialize MCP servers
	if err := mcp.LoadConfig(rootPath); err != nil {
		fmt.Println(utils.ErrorMessage(fmt.Sprintf("Failed to load MCP servers: %s", err.Error())))
		os.Exit(1)
	}

	// If MCP initialization is enabled in the configuration, proceed with initialization
	if app.OclaiConfig.InitMCP {
		if err := mcp.InitializeServers(ctx, rootPath); err != nil {
			fmt.Println(utils.ErrorMessage(fmt.Sprintf("Failed to initialize MCP servers: %s", err.Error())))
			os.Exit(1)
		}

		// Disable MCP initialization after successful initialization
		app.OclaiConfig.InitMCP = false

		// Update the configuration file with the new settings
		if err = app.UpdateConfig(rootPath); err != nil {
			fmt.Println(utils.ErrorMessage(err.Error()))
			os.Exit(1)
		}
	}

	// Execute the command-line interface
	cmd.Execute()
}
