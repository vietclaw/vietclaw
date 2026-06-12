package config

import "fmt"

func Validate(cfg Config) error {
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535")
	}
	if cfg.Runtime.MaxConcurrentTasks < 1 {
		return fmt.Errorf("runtime.max_concurrent_tasks must be >= 1")
	}
	if cfg.Agent.MaxSteps < 0 {
		return fmt.Errorf("agent.max_steps must be >= 0")
	}
	if cfg.Agent.MaxOutputTokens < 0 {
		return fmt.Errorf("agent.max_output_tokens must be >= 0")
	}
	if cfg.Budget.DailyUSDLimit < 0 || cfg.Budget.RequireApprovalAboveUSD < 0 {
		return fmt.Errorf("budget values must be >= 0")
	}
	if cfg.Tools.Shell.Sandbox != "" && cfg.Tools.Shell.Sandbox != "none" && cfg.Tools.Shell.Sandbox != "docker" {
		return fmt.Errorf("tools.shell.sandbox must be none or docker")
	}
	if cfg.Tools.Shell.WorkspaceMode != "" && cfg.Tools.Shell.WorkspaceMode != "ro" && cfg.Tools.Shell.WorkspaceMode != "rw" {
		return fmt.Errorf("tools.shell.workspace_mode must be ro or rw")
	}
	if cfg.Tools.Shell.TimeoutSeconds < 0 {
		return fmt.Errorf("tools.shell.timeout_seconds must be >= 0")
	}
	for _, provider := range cfg.Providers {
		if provider.ID == "" {
			return fmt.Errorf("provider id is required")
		}
		if provider.Type == "" {
			return fmt.Errorf("provider %s type is required", provider.ID)
		}
	}
	if cfg.Framework.MaxTotalAgents < 1 {
		return fmt.Errorf("framework.max_total_agents must be >= 1")
	}
	if cfg.Framework.MaxConcurrentSpawns < 1 {
		return fmt.Errorf("framework.max_concurrent_spawns must be >= 1")
	}
	for _, entry := range cfg.Models.Catalog {
		if entry.ID == "" {
			return fmt.Errorf("models.catalog entry id is required")
		}
		if entry.Provider == "" {
			return fmt.Errorf("models.catalog entry %s provider is required", entry.ID)
		}
		if entry.Model == "" {
			return fmt.Errorf("models.catalog entry %s model is required", entry.ID)
		}
	}
	if cfg.Channels.Telegram.CommandMode != "" &&
		cfg.Channels.Telegram.CommandMode != "slash" &&
		cfg.Channels.Telegram.CommandMode != "prefix" {
		return fmt.Errorf("channels.telegram.command_mode must be slash or prefix")
	}
	return nil
}
