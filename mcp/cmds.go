package mcp

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thejasmeetsingh/oclai/utils"
)

// rootPath stores the root directory path for MCP server configurations
var rootPath = ""

var (
	// McpRootCmd is the root command for MCP server management
	McpRootCmd = &cobra.Command{
		Use:   "mcp",
		Short: "Manage MCP servers",
		Long:  utils.InfoBox("Manage MCP servers. This command allows you to list, add, and remove MCP servers with various configurations."),
		Example: `
		oclai mcp ls
		oclai mcp add --name everything --cmd npx --args '-y @modelcontextprotocol/server-everything'
		oclai mcp rm everything
	`,
	}

	// listServersCmd lists available MCP servers
	listServersCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List MCP servers",
		Long:    utils.InfoBox("List MCP servers. This command displays all available MCP servers in a formatted list."),
		Example: "oclai mcp ls",
		Run: func(cmd *cobra.Command, args []string) {
			// Get the list of servers from storage
			servers := getServerList()

			// If no servers are available, show an error message
			if len(servers) == 0 {
				fmt.Println(utils.ErrorBox("No servers are available. Please add a server 🌫️"))
				os.Exit(0)
			}

			// Build the result string with server list
			result := "# Available Servers\n"

			for _, server := range servers {
				result += fmt.Sprintf("- %s\n", server)
			}

			// Convert the result to markdown format
			md, err := utils.ToMarkDown(result)
			if err != nil {
				fmt.Println(utils.ErrorMessage(fmt.Sprintf("Error caught while converting to markdown: %s", err)))
				os.Exit(1)
			}

			fmt.Println(md)
		},
	}

	// removeServerCmd removes a specific MCP server
	removeServerCmd = &cobra.Command{
		Use:     "remove [name]",
		Aliases: []string{"rm"},
		Short:   "Remove a MCP server",
		Long:    utils.InfoBox("Remove a MCP server. This command allows you to remove a server by specifying its name."),
		Example: "oclai mcp rm everything",
		Run: func(cmd *cobra.Command, args []string) {
			// Extract the server name from command arguments
			serverName := strings.TrimSpace(strings.Join(args, " "))

			// Validate that a server name was provided
			if serverName == "" {
				fmt.Println(utils.ErrorMessage("Please provide the server name 😒"))
				os.Exit(1)
			}

			// Remove the server from storage
			err := removeServer(rootPath, serverName)
			if err != nil {
				fmt.Println(utils.ErrorMessage(fmt.Sprintf("Error caught while removing a server: %s", err)))
				os.Exit(1)
			}

			fmt.Println(utils.SuccessBox(fmt.Sprintf("'%s' server removed successfully!", serverName)))
		},
	}

	// addServerCmd adds a new MCP server with specified configurations
	addServerCmd = &cobra.Command{
		Use:   "add",
		Short: "Add a MCP server",
		Long:  utils.InfoBox("Add a MCP server. This command allows you to add a new server with various configurations such as command, endpoint, headers, and environment variables."),
		Example: `
		oclai mcp add --name everything --cmd npx --args '-y @modelcontextprotocol/server-everything'
		oclai mcp add --name brave-search --cmd docker --args 'run -i --rm mcp/brave-search' --env=BRAVE_API_KEY:$BRAVE_API_KEY
		oclai mcp add --name github --endpoint https://api.githubcopilot.com/mcp/ --headers=Authorization:Bearer $GH_TOKEN
		`,
		Run: func(cmd *cobra.Command, args []string) {
			// Retrieve command line flags
			nameArg, _ := cmd.Flags().GetString("name")
			cmdArg, _ := cmd.Flags().GetString("cmd")
			endpointArg, _ := cmd.Flags().GetString("endpoint")
			cmdArgs, _ := cmd.Flags().GetString("args")
			headerArgs, _ := cmd.Flags().GetStringSlice("headers")
			envArgs, _ := cmd.Flags().GetStringSlice("env")

			var (
				isSSE   bool
				headers map[string]string
				env     map[string]string
			)

			// Validate that a server name was provided
			if nameArg == "" {
				fmt.Println(utils.ErrorMessage("'--name' is required 🤌"))
				os.Exit(1)
			}

			// Validate that either command or endpoint was provided
			if cmdArg == "" && endpointArg == "" {
				fmt.Println(utils.ErrorMessage("'--cmd' or '--endpoint' is required 🤌"))
				os.Exit(1)
			}

			// Prevent using both command and endpoint together
			if cmdArg != "" && endpointArg != "" {
				fmt.Println(utils.ErrorMessage("Cannot add '--cmd' and '--endpoint' together 🤝"))
				os.Exit(1)
			}

			// Handle command-based server configuration
			if cmdArg != "" {
				isSSE = false
				env = getArrayToMap(envArgs)

				// Process command arguments
				if cmdArgs != "" {
					args = strings.Split(strings.TrimSpace(cmdArgs), " ")
				}
			} else {
				// Handle endpoint-based server configuration
				isSSE = true
				headers = getArrayToMap(headerArgs)
			}

			// Add MCP server
			err := addServer(rootPath, McpServer{
				IsSSE:    isSSE,
				Name:     strings.TrimSpace(nameArg),
				Command:  strings.TrimSpace(cmdArg),
				Args:     args,
				Endpoint: strings.TrimSpace(endpointArg),
				Headers:  headers,
				Env:      env,
			})

			if err != nil {
				fmt.Println(utils.ErrorMessage(fmt.Sprintf("Error caught while adding the server: %s", err)))
				os.Exit(1)
			}

			fmt.Println(utils.SuccessBox("Server added successfully!"))
		},
	}
)

// A helper function to convert cmd args - From an array to map
func getArrayToMap(arr []string) map[string]string {
	result := make(map[string]string)

	for _, el := range arr {
		keyVal := strings.Split(el, ":")
		key := strings.TrimSpace(keyVal[0])
		val := strings.TrimSpace(keyVal[1])

		result[key] = val
	}

	return result
}

func init() {
	// Get application root directory
	_rootPath, err := utils.GetAppRootDir()
	if err != nil {
		fmt.Println(utils.ErrorMessage(fmt.Sprintf("Error caught while retreiving root path: %s", err)))
		os.Exit(1)
	}

	rootPath = _rootPath

	// Add sub-commands to mcp root cmd
	McpRootCmd.AddCommand(listServersCmd, addServerCmd, removeServerCmd)

	// Register add mcp server command flags
	addServerCmd.Flags().StringP("name", "n", "", "Server name")
	addServerCmd.Flags().String("cmd", "", "Command to start the server")
	addServerCmd.Flags().String("endpoint", "", "HTTP/SSE endpoint of the server")
	addServerCmd.Flags().String("args", "", "Arguments for the server command")
	addServerCmd.Flags().StringSlice("env", []string{}, "Specify env varriables (comma seperated) to run the server command with")
	addServerCmd.Flags().StringSlice("headers", []string{}, "Add addition headers varriables (comma seperated) which will be used while connecting to the server")
}
