package framework

import (
	"context"
	"log"

	"vietclaw/internal/config"
)

// Framework is the extension and lifecycle layer above the agent runtime.
type Framework struct {
	Config config.FrameworkConfig
	Hooks  *HookRegistry
	Logger *log.Logger
}

func New(cfg config.FrameworkConfig, logger *log.Logger) *Framework {
	f := &Framework{
		Config: cfg,
		Hooks:  NewHookRegistry(),
		Logger: logger,
	}
	if cfg.HooksEnabled {
		f.registerBuiltinHooks()
	}
	return f
}

func (f *Framework) registerBuiltinHooks() {
	f.Hooks.Register(EventBeforeTool, func(ctx context.Context, hc HookContext) error {
		if f.Logger != nil && hc.ToolName != "" {
			f.Logger.Printf("framework hook before_tool session=%s tool=%s", hc.SessionID, hc.ToolName)
		}
		return nil
	})
	f.Hooks.Register(EventRunFinish, func(ctx context.Context, hc HookContext) error {
		if f.Logger != nil && hc.RunID != "" {
			f.Logger.Printf("framework hook run_finish run=%s agent=%s", hc.RunID, hc.AgentID)
		}
		return nil
	})
}

type ExtensionInfo struct {
	Kind    string   `json:"kind"`
	Builtin []string `json:"builtin"`
}

func BuiltinExtensions() []ExtensionInfo {
	return []ExtensionInfo{
		{Kind: "tools", Builtin: []string{
			"file_read", "file_write", "shell_exec", "web_search", "web_fetch",
			"memory_recall", "memory_store", "agent_delegate", "mcp",
		}},
		{Kind: "channels", Builtin: []string{"discord", "telegram"}},
		{Kind: "providers", Builtin: []string{
			"mock", "openai", "openai-compatible", "anthropic", "gemini", "http", "opencode-zen",
		}},
		{Kind: "hooks", Builtin: []string{
			string(EventBeforeChat), string(EventAfterChat),
			string(EventBeforeTool), string(EventAfterTool),
			string(EventRunStart), string(EventRunFinish),
		}},
	}
}
