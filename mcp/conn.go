package mcp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	goMCP "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/thejasmeetsingh/oclai/ollama"
)

type (
	ConnParams struct {
		IsSSE    bool
		Command  string
		Args     []string
		Endpoint string
		Headers  map[string]string
		envVars  map[string]string
	}

	customTransport struct {
		headers             map[string]string
		underlyingTransport http.RoundTripper
	}
)

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

	session, err := Client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func InitializeServers(ctx context.Context, servers map[string]map[string]any) error {
	var allTools map[string][]ollama.Tool

	for server, srvConfig := range servers {
		command, isCommand := srvConfig["command"]
		args, isArgs := srvConfig["args"]
		endpoint, isEndpoint := srvConfig["endpoint"]
		headers, isHeaders := srvConfig["headers"]
		envVars, isEnvVars := srvConfig["env"]

		if !isCommand && !isEndpoint {
			return fmt.Errorf("no transport is provided for %s server", server)
		}

		if !isArgs {
			args = make([]string, 0)
		}

		if !isHeaders {
			headers = make(map[string]string)
		}

		if !isEnvVars {
			envVars = make(map[string]string)
		}

		var connParams ConnParams

		if isCommand {
			connParams.IsSSE = false
			connParams.Command = command.(string)
			connParams.Args = args.([]string)
			connParams.envVars = envVars.(map[string]string)
		} else {
			connParams.IsSSE = true
			connParams.Endpoint = endpoint.(string)
			connParams.Headers = headers.(map[string]string)
		}

		session, err := CreateSession(ctx, connParams)
		if err != nil {
			return err
		}

		tools, err := ListTools(ctx, session)
		if err != nil {
			return err
		}

		if len(tools) != 0 {
			allTools[server] = tools
		}
	}

	err := SaveTools(allTools)
	if err != nil {
		return err
	}

	return nil
}
