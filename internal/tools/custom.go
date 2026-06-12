package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"vietclaw/internal/agentfs"
	"vietclaw/internal/config"
)

type customToolRunner struct {
	registry *agentfs.Registry
	agentID  string
	tool     agentfs.CustomTool
	policy   Policy
	cfg      config.Config
	mcp      map[string]mcpToolRef
}

func (c customToolRunner) Name() string {
	return c.tool.Name
}

func (c customToolRunner) Run(ctx context.Context, argsJSON string) (string, error) {
	handler := strings.TrimSpace(c.tool.Handler)
	switch {
	case strings.HasPrefix(handler, "mcp:"):
		serverID := strings.TrimPrefix(handler, "mcp:")
		prefix := mcpToolPrefix + "_" + sanitizeToolName(serverID) + "_"
		for name, ref := range c.mcp {
			if strings.HasPrefix(name, prefix) {
				return ref.client.Execute(ctx, ref.name, argsJSON)
			}
		}
		return "", fmt.Errorf("no MCP tools found for server %s", serverID)
	case strings.HasPrefix(handler, "script:"):
		if !c.cfg.Tools.Shell.Enabled {
			return "", fmt.Errorf("script custom tools require shell.enabled")
		}
		script := strings.TrimSpace(strings.TrimPrefix(handler, "script:"))
		if !filepath.IsAbs(script) {
			def, ok := c.registry.Get(c.agentID)
			if !ok {
				return "", fmt.Errorf("agent not found")
			}
			script = filepath.Join(def.Dir, script)
		}
		timeout := time.Duration(c.cfg.Tools.Shell.TimeoutSeconds) * time.Second
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		cmd := exec.CommandContext(ctx, script)
		cmd.Stdin = strings.NewReader(argsJSON)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return string(out), fmt.Errorf("script failed: %w", err)
		}
		return strings.TrimSpace(string(out)), nil
	default:
		if strings.TrimSpace(c.tool.Instructions) != "" {
			var args map[string]any
			_ = json.Unmarshal([]byte(argsJSON), &args)
			return fmt.Sprintf("%s\nArgs: %v", c.tool.Instructions, args), nil
		}
		return "", fmt.Errorf("custom tool handler not configured for %s", c.tool.Name)
	}
}

func (c customToolRunner) WorkspacePath() string {
	if c.cfg.Agent.Workspace == "" {
		return ""
	}
	return c.cfg.Agent.Workspace
}
