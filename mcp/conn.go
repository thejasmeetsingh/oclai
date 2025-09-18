package mcp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	goMCP "github.com/modelcontextprotocol/go-sdk/mcp"
)

type customTransport struct {
	headers             map[string]string
	underlyingTransport http.RoundTripper
}

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.headers != nil {
		for key, value := range t.headers {
			req.Header.Add(key, value)
		}
	}

	return t.underlyingTransport.RoundTrip(req)
}

func setEnvVars(vars map[string]string) error {
	for key, value := range vars {
		err := os.Setenv(key, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateSession(ctx context.Context, server McpServer) (*goMCP.ClientSession, error) {
	if len(server.Env) != 0 {
		err := setEnvVars(server.Env)
		if err != nil {
			return nil, err
		}
	}

	var transport goMCP.Transport

	if server.IsSSE {
		httpClient := http.Client{}
		httpClient.Transport = &customTransport{
			headers:             server.Headers,
			underlyingTransport: http.DefaultTransport,
		}

		transport = &goMCP.SSEClientTransport{Endpoint: server.Endpoint, HTTPClient: &httpClient}
	} else {
		transport = &goMCP.CommandTransport{Command: exec.Command(server.Command, server.Args...)}
	}

	session, err := Client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func InitializeServers(ctx context.Context, rootPath string) error {
	servers := McpServers["servers"]

	for _, server := range servers {
		if server.Command == "" && server.Endpoint == "" {
			return fmt.Errorf("no transport is provided for %s server", server.Name)
		}

		if len(server.Args) == 0 {
			server.Args = make([]string, 0)
		}

		if len(server.Headers) == 0 {
			server.Headers = make(map[string]string)
		}

		if len(server.Env) == 0 {
			server.Env = make(map[string]string)
		}

		session, err := CreateSession(ctx, *server)
		if err != nil {
			return err
		}

		tools, err := ListTools(ctx, session)
		if err != nil {
			return err
		}

		if len(tools) != 0 {
			server.Tools = tools
		}

		session.Close()
	}

	err := UpdateConfig(rootPath)
	if err != nil {
		return err
	}

	return nil
}
