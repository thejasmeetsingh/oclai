package mcp

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/thejasmeetsingh/oclai/utils"
)

var (
	rootPath = ""

	infoMsg    = color.New(color.FgBlue, color.Bold)
	errMsg     = color.New(color.FgRed)
	successMsg = color.New(color.FgGreen)
)

var (
	McpRootCmd = &cobra.Command{
		Use:   "mcp",
		Short: "Manage MCP servers",
		Long:  infoMsg.Sprint("Manage MCP servers. This command allows you to list, add, and remove MCP servers with various configurations."),
		Example: `
		oclai mcp ls
		oclai mcp add --name everything --cmd npx --args '-y @modelcontextprotocol/server-everything'
		oclai mcp rm everything
	`,
	}

	listServersCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List MCP servers",
		Long:    infoMsg.Sprint("List MCP servers. This command displays all available MCP servers in a formatted list."),
		Example: "oclai mcp ls",
		Run: func(cmd *cobra.Command, args []string) {
			servers := getServerList()

			if len(servers) == 0 {
				infoMsg.Println("No servers are available. Please add a server 🌫️")
				os.Exit(0)
			}

			result := "# Available Servers\n"

			for _, server := range servers {
				result += fmt.Sprintf("- %s\n", server)
			}

			md, err := utils.ToMarkDown(result)
			if err != nil {
				errMsg.Println("Error caught while converting to markdown: ", err)
				os.Exit(1)
			}

			fmt.Println(md)
		},
	}

	removeServerCmd = &cobra.Command{
		Use:     "remove [name]",
		Aliases: []string{"rm"},
		Short:   "Remove a MCP server",
		Long:    infoMsg.Sprint("Remove a MCP server. This command allows you to remove a server by specifying its name."),
		Example: "oclai mcp rm everything",
		Run: func(cmd *cobra.Command, args []string) {
			serverName := strings.TrimSpace(strings.Join(args, " "))

			if serverName == "" {
				errMsg.Println("Please provide the server name 😒")
				os.Exit(1)
			}

			err := removeServer(rootPath, serverName)
			if err != nil {
				errMsg.Println("Error caught while removing a server: ", err)
				os.Exit(1)
			}

			successMsg.Printf("'%s' server removed successfully!\n", serverName)
		},
	}

	addServerCmd = &cobra.Command{
		Use:   "add",
		Short: "Add a MCP server",
		Long:  infoMsg.Sprint("Add a MCP server. This command allows you to add a new server with various configurations such as command, endpoint, headers, and environment variables."),
		Example: `
		oclai mcp add --name everything --cmd npx --args '-y @modelcontextprotocol/server-everything'
		oclai mcp add --name brave-search --cmd docker --args 'run -i --rm mcp/brave-search' --env=BRAVE_API_KEY:$BRAVE_API_KEY
		oclai mcp add --name github --endpoint https://api.githubcopilot.com/mcp/ --headers=Authorization:Bearer $GH_TOKEN
		`,
		Run: func(cmd *cobra.Command, args []string) {
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

			if nameArg == "" {
				errMsg.Println("'--name' is required 🤌")
				os.Exit(1)
			}

			if cmdArg == "" && endpointArg == "" {
				errMsg.Println("'--cmd' or '--endpoint' is required 🤌")
				os.Exit(1)
			}

			if cmdArg != "" && endpointArg != "" {
				errMsg.Println("Cannot add '--cmd' and '--endpoint' together 🤝")
				os.Exit(1)
			}

			if cmdArg != "" {
				isSSE = false
				env = getArrayToMap(envArgs)

				if cmdArgs != "" {
					args = strings.Split(strings.TrimSpace(cmdArgs), " ")
				}
			} else {
				isSSE = true
				headers = getArrayToMap(headerArgs)
			}

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
				errMsg.Println("Error caught while adding the server: ", err)
				os.Exit(1)
			}

			successMsg.Println("Server added successfully!")
		},
	}
)

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
	_rootPath, err := utils.GetAppRootDir()
	if err != nil {
		errMsg.Println("Error caught while retreiving root path: ", err)
		os.Exit(1)
	}

	rootPath = _rootPath

	McpRootCmd.AddCommand(listServersCmd)
	McpRootCmd.AddCommand(addServerCmd)
	McpRootCmd.AddCommand(removeServerCmd)

	addServerCmd.Flags().StringP("name", "n", "", "Server name")
	addServerCmd.Flags().String("cmd", "", "Command to start the server")
	addServerCmd.Flags().String("endpoint", "", "HTTP/SSE endpoint of the server")
	addServerCmd.Flags().String("args", "", "Arguments for the server command")
	addServerCmd.Flags().StringSlice("env", []string{}, "Specify env varriables (comma seperated) to run the server command with")
	addServerCmd.Flags().StringSlice("headers", []string{}, "Add addition headers varriables (comma seperated) which will be used while connecting to the server")
}
