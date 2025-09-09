package mcp

import (
	"context"
	"net/http"
	"os"
	"os/exec"

	goMCP "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/thejasmeetsingh/oclai/app"
)

var client = app.McpClient

type (
	ConnParams struct {
		IsSSE    bool
		Command  string
		Args     []string
		Endpoint string
		Headers  *map[string]string
		envVars  *map[string]string
	}

	customTransport struct {
		headers             *map[string]string
		underlyingTransport http.RoundTripper
	}
)

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.headers != nil {
		for key, value := range *t.headers {
			req.Header.Add(key, value)
		}
	}

	return t.underlyingTransport.RoundTrip(req)
}

func setEnvVars(vars *map[string]string) error {
	for key, value := range *vars {
		err := os.Setenv(key, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateSession(ctx context.Context, params ConnParams) (*goMCP.ClientSession, error) {
	if params.envVars != nil {
		err := setEnvVars(params.envVars)
		if err != nil {
			return nil, err
		}
	}

	var transport goMCP.Transport

	if params.IsSSE {
		httpClient := http.Client{}
		httpClient.Transport = &customTransport{
			headers:             params.Headers,
			underlyingTransport: http.DefaultTransport,
		}

		transport = &goMCP.SSEClientTransport{Endpoint: params.Endpoint, HTTPClient: &httpClient}
	} else {
		transport = &goMCP.CommandTransport{Command: exec.Command(params.Command, params.Args...)}
	}

	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, err
	}

	return session, nil
}
