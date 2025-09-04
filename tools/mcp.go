package tools

import (
	"context"
	"fmt"
	"log"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var client = mcp.NewClient(&mcp.Implementation{Name: "oclai", Version: "v1.0.0"}, nil)

func ListTools() {
	ctx := context.Background()

	clientSession, err := client.Connect(ctx, &mcp.CommandTransport{Command: exec.Command("npx", "-y", "@modelcontextprotocol/server-everything")}, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer clientSession.Close()

	// mcpTools := clientSession.Tools(ctx, nil)

	// for mcpTool, err := range mcpTools {
	// 	if err != nil {
	// 		fmt.Println(err.Error())
	// 		continue
	// 	}
	// 	fmt.Println(mcpTool.Name + " => " + mcpTool.Description)
	// 	if mcpTool.InputSchema != nil {
	// 		fmt.Println("Input Schema:")
	// 		inputSchema, _ := mcpTool.InputSchema.MarshalJSON()
	// 		fmt.Println(string(inputSchema))
	// 	}
	// 	if mcpTool.OutputSchema != nil {
	// 		fmt.Println("Output Schema:")
	// 		outputSchema, _ := mcpTool.OutputSchema.MarshalJSON()
	// 		fmt.Println(string(outputSchema))
	// 	}
	// 	fmt.Println(strings.Repeat("---", 200))
	// }

	result, _ := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "structuredContent",
		Arguments: map[string]any{
			"location": "New Delhi",
		},
	})
	for _, content := range result.Content {
		data, err := content.MarshalJSON()
		if err != nil {
			fmt.Println("Error: ", err.Error())
		}
		fmt.Println(string(data))
	}
}
