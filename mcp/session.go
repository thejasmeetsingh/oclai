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

func CreateSession(ctx context.Context, server McpServer) (*goMCP.ClientSession, error) {
	var transport goMCP.Transport

	if server.IsSSE {
		httpClient := http.Client{}
		httpClient.Transport = &customTransport{
			headers:             server.Headers,
			underlyingTransport: http.DefaultTransport,
		}

		transport = &goMCP.SSEClientTransport{Endpoint: server.Endpoint, HTTPClient: &httpClient}
	} else {
		cmd := exec.Command(server.Command, server.Args...)

		if len(server.Env) != 0 {
			isDockerCmd := server.Command == "docker"
			env := getEnv(server.Env, isDockerCmd)

			if len(env) != 0 {
				if isDockerCmd {
					args := server.Args
					args = append(args, env...)
					cmd.Args = args
				} else {
					cmd.Env = env
				}
			}
		}

		transport = &goMCP.CommandTransport{Command: cmd}
	}

	session, err := Client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, err
	}

	return session, nil
}
