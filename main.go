package main

import (
	"context"
	"fmt"
	"os"

	"github.com/thejasmeetsingh/oclai/app"
	"github.com/thejasmeetsingh/oclai/cmd"
	"github.com/thejasmeetsingh/oclai/mcp"
	"github.com/thejasmeetsingh/oclai/utils"
)

func main() {
	ctx := context.Background()

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

	if app.OclaiConfig.InitMCP {
		if err := mcp.InitializeServers(ctx, rootPath); err != nil {
			fmt.Println(utils.ErrorMessage(fmt.Sprintf("Failed to initialize MCP servers: %s", err.Error())))
			os.Exit(1)
		}

		app.OclaiConfig.InitMCP = false
		if err = app.UpdateConfig(rootPath); err != nil {
			fmt.Println(utils.ErrorMessage(err.Error()))
			os.Exit(1)
		}
	}

	cmd.Execute()
}
