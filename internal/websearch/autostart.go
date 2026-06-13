package websearch

import (
	"vietclaw/internal/config"
)

// EnsureMCPEnabled registers the open-websearch MCP server in cfg when missing.
// Returns the updated config and whether a new entry was added.
func EnsureMCPEnabled(cfg config.Config) (config.Config, bool, error) {
	if _, ok := Find(cfg); ok {
		return cfg, false, nil
	}
	server, err := MCPServerConfig("", nil)
	if err != nil {
		return cfg, false, err
	}
	return UpsertMCPServer(cfg, server), true, nil
}
