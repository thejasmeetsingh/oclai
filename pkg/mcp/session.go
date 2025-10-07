package mcp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	goMCP "github.com/modelcontextprotocol/go-sdk/mcp"
)

// customTransport is a custom HTTP transport that allows adding custom headers to requests.
// It wraps the underlying transport and adds headers to the request before sending.
type customTransport struct {
	headers             map[string]string
	underlyingTransport http.RoundTripper
}

// RoundTrip implements the http.RoundTripper interface.
// It adds custom headers to the request if they are defined, then delegates to the underlying transport.
func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.headers != nil {
		for key, value := range t.headers {
			req.Header.Add(key, value)
		}
	}

	return t.underlyingTransport.RoundTrip(req)
}

// getEnv processes an environment map and returns a slice of strings suitable for passing to a command.
// It handles environment variables that start with $ by expanding them using os.Getenv.
// It also formats the environment variables according to whether the command is Docker or not.
func getEnv(env map[string]string, isDockerCmd bool) []string {
	result := make([]string, 0)

	for key, val := range env {
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)

		if val[0] == '$' {
			val = os.Getenv(val[1:])
		}

		if val != "" {
			envVar := fmt.Sprintf("%s=%s", key, val)
			if isDockerCmd {
				envVar = "-e " + envVar
			}

			result = append(result, envVar)
		}
	}

	return result
}

// createSession creates and returns a new MCP client session based on the server configuration.
// It handles both SSE (Server-Sent Events) and command-based server types.
func createSession(ctx context.Context, server McpServer) (*goMCP.ClientSession, error) {
	var transport goMCP.Transport

	if server.IsSSE {
		// For SSE, we create an HTTP client with custom headers and use the StreamableClientTransport.
		httpClient := http.Client{}
		httpClient.Transport = &customTransport{
			headers:             server.Headers,
			underlyingTransport: http.DefaultTransport,
		}
		transport = &goMCP.StreamableClientTransport{Endpoint: server.Endpoint, HTTPClient: &httpClient}
	} else {
		// For command-based servers, we create an exec.Cmd and use the CommandTransport.
		cmd := exec.Command(server.Command, server.Args...)
		cmd.Env = os.Environ()

		if len(server.Env) != 0 {
			isDockerCmd := server.Command == "docker"
			env := getEnv(server.Env, isDockerCmd)

			if len(env) != 0 {
				if isDockerCmd {
					// If the command is Docker, append the environment variables as arguments.
					args := server.Args
					args = append(args, env...)
					cmd.Args = args
				} else {
					// Otherwise, set the environment variables directly on the command.
					cmd.Env = append(cmd.Env, env...)
				}
			}
		}

		transport = &goMCP.CommandTransport{Command: cmd}
	}

	// Create and return the client session.
	session, err := Client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, err
	}

	return session, nil
}
