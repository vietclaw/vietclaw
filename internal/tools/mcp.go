package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"vietclaw/internal/config"
	"vietclaw/internal/providers"
)

const (
	mcpToolPrefix        = "mcp"
	defaultMCPHTTPClient = 10 * time.Second
)

type MCPClient struct {
	server config.MCPServerConfig
	client *http.Client
}

func NewMCPClient(server config.MCPServerConfig) *MCPClient {
	return &MCPClient{
		server: server,
		client: &http.Client{Timeout: defaultMCPHTTPClient},
	}
}

type MCPDiscoveredTool struct {
	Name       string
	Definition providers.ToolDefinition
}

func (c *MCPClient) Discover(ctx context.Context) ([]MCPDiscoveredTool, error) {
	var result struct {
		Tools []struct {
			Name        string         `json:"name"`
			Description string         `json:"description"`
			InputSchema map[string]any `json:"inputSchema"`
		} `json:"tools"`
	}
	if err := c.call(ctx, "tools/list", map[string]any{}, &result); err != nil {
		return nil, err
	}
	discovered := make([]MCPDiscoveredTool, 0, len(result.Tools))
	for _, tool := range result.Tools {
		name := mcpToolName(c.server.ID, tool.Name)
		discovered = append(discovered, MCPDiscoveredTool{
			Name: tool.Name,
			Definition: providers.ToolDefinition{
				Type: "function",
				Function: providers.FunctionDetail{
					Name:        name,
					Description: tool.Description,
					Parameters:  tool.InputSchema,
				},
			},
		})
	}
	return discovered, nil
}

func (c *MCPClient) Tools(ctx context.Context) ([]providers.ToolDefinition, error) {
	discovered, err := c.Discover(ctx)
	if err != nil {
		return nil, err
	}
	defs := make([]providers.ToolDefinition, 0, len(discovered))
	for _, tool := range discovered {
		defs = append(defs, tool.Definition)
	}
	return defs, nil
}

func (c *MCPClient) Execute(ctx context.Context, toolName, argsJSON string) (string, error) {
	var args map[string]any
	if strings.TrimSpace(argsJSON) != "" {
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return "", err
		}
	}
	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := c.call(ctx, "tools/call", map[string]any{
		"name":      toolName,
		"arguments": args,
	}, &result); err != nil {
		return "", err
	}
	var parts []string
	for _, item := range result.Content {
		if item.Text != "" {
			parts = append(parts, item.Text)
		}
	}
	return strings.Join(parts, "\n"), nil
}

func (c *MCPClient) call(ctx context.Context, method string, params map[string]any, out any) error {
	if c.server.URL == "" {
		return fmt.Errorf("mcp server %s missing url", c.server.ID)
	}
	body, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  method,
		"params":  params,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.server.URL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("mcp server %s returned %s", c.server.ID, resp.Status)
	}

	var payload struct {
		Result json.RawMessage `json:"result"`
		Error  struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return err
	}
	if payload.Error.Message != "" {
		return fmt.Errorf("mcp server %s error: %s", c.server.ID, payload.Error.Message)
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(payload.Result, out)
}

func mcpToolName(serverID, toolName string) string {
	return mcpToolPrefix + "_" + sanitizeToolName(serverID) + "_" + sanitizeToolName(toolName)
}

func sanitizeToolName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer(".", "_", "-", "_", " ", "_", "/", "_", ":", "_")
	return replacer.Replace(value)
}
