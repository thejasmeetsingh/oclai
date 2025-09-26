package main

import (
	"context"
	"os"

	"github.com/fatih/color"
	"github.com/thejasmeetsingh/oclai/app"
	"github.com/thejasmeetsingh/oclai/cmd"
	"github.com/thejasmeetsingh/oclai/mcp"
	"github.com/thejasmeetsingh/oclai/utils"
)

var errMsg = color.New(color.FgRed)

func main() {
	ctx := context.Background()

	rootPath, err := utils.GetAppRootDir()
	if err != nil {
		errMsg.Println("Error caught while retreiving root path: ", err)
		os.Exit(1)
	}

	// Load App configuration file
	if err := app.LoadConfig(rootPath); err != nil {
		errMsg.Println("Failed to load configuration: ", err)
		os.Exit(1)
	}

	// Load and Initialize MCP servers
	if err := mcp.LoadConfig(rootPath); err != nil {
		errMsg.Println("Failed to load MCP servers: ", err)
		os.Exit(1)
	}

	if app.OclaiConfig.InitMCP {
		if err := mcp.InitializeServers(ctx, rootPath); err != nil {
			errMsg.Println("Failed to initialize MCP servers: ", err)
			os.Exit(1)
		}

		app.OclaiConfig.InitMCP = false
		if err = app.UpdateConfig(rootPath); err != nil {
			errMsg.Println(err)
			os.Exit(1)
		}
	}

	cmd.Execute()
}
