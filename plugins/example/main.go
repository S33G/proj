package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

// RPC types
type RPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      int             `json:"id"`
}

type RPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
	ID      int         `json:"id"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Plugin types
type Project struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Language  string `json:"language"`
	GitBranch string `json:"gitBranch"`
	GitDirty  bool   `json:"gitDirty"`
	IsGitRepo bool   `json:"isGitRepo"`
}

type Action struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Priority    int    `json:"priority"`
}

type ActionResult struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	CdPath  string   `json:"cdPath,omitempty"`
	ExecCmd []string `json:"execCmd,omitempty"`
}

// ExamplePlugin is a simple example plugin
type ExamplePlugin struct {
	config map[string]interface{}
}

func main() {
	plugin := &ExamplePlugin{}
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Bytes()

		var req RPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to unmarshal request: %v\n", err)
			continue
		}

		resp := plugin.handleRequest(&req)

		respData, err := json.Marshal(resp)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to marshal response: %v\n", err)
			continue
		}

		fmt.Println(string(respData))
	}
}

func (p *ExamplePlugin) handleRequest(req *RPCRequest) *RPCResponse {
	resp := &RPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	case "init":
		var params struct {
			Config map[string]interface{} `json:"config"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			resp.Error = &RPCError{Code: -32602, Message: "Invalid params"}
			return resp
		}
		p.config = params.Config
		resp.Result = map[string]bool{"success": true}

	case "actions":
		var proj Project
		if err := json.Unmarshal(req.Params, &proj); err != nil {
			resp.Error = &RPCError{Code: -32602, Message: "Invalid params"}
			return resp
		}
		resp.Result = p.getActions(proj)

	case "executeAction":
		var params struct {
			Action  string  `json:"action"`
			Project Project `json:"project"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			resp.Error = &RPCError{Code: -32602, Message: "Invalid params"}
			return resp
		}
		resp.Result = p.executeAction(params.Action, params.Project)

	case "shutdown":
		resp.Result = map[string]bool{"success": true}
		os.Exit(0)

	default:
		resp.Error = &RPCError{Code: -32601, Message: "Method not found"}
	}

	return resp
}

func (p *ExamplePlugin) getActions(proj Project) []Action {
	return []Action{
		{
			ID:          "example-hello",
			Label:       "Say Hello",
			Description: "Example plugin action",
			Icon:        "👋",
			Priority:    100,
		},
	}
}

func (p *ExamplePlugin) executeAction(actionID string, proj Project) ActionResult {
	if actionID == "example-hello" {
		return ActionResult{
			Success: true,
			Message: fmt.Sprintf("Hello from example plugin!\nProject: %s", proj.Name),
		}
	}

	return ActionResult{
		Success: false,
		Message: "Unknown action",
	}
}
