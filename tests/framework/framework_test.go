package framework_test

import (
	"context"
	"testing"

	"vietclaw/internal/channels"
	_ "vietclaw/internal/channels/discord"
	_ "vietclaw/internal/channels/telegram"
	"vietclaw/internal/config"
	"vietclaw/internal/framework"
	"vietclaw/internal/tools"
)

func TestFrameworkDefaults(t *testing.T) {
	cfg := config.Default(config.Paths{DataDir: t.TempDir()})
	if !cfg.Framework.Enabled || !cfg.Framework.DelegateEnabled || !cfg.Framework.HooksEnabled {
		t.Fatalf("framework defaults invalid: %#v", cfg.Framework)
	}
}

func TestChannelRegistry(t *testing.T) {
	got := channels.RegisteredAdapters()
	if len(got) < 2 {
		t.Fatalf("adapters = %#v", got)
	}
}

func TestToolRegisterAndFilter(t *testing.T) {
	cfg := config.Default(config.Paths{DataDir: t.TempDir()})
	reg := tools.NewRegistry(cfg)
	reg.Register(tools.SystemInfo{}, tools.AgentDelegateDefinition())
	defs := reg.GetDefinitionsForProfile(config.AgentProfileConfig{Tools: []string{"system_info"}}, false)
	if len(defs) != 1 || defs[0].Function.Name != "system_info" {
		t.Fatalf("filtered defs = %#v", defs)
	}
}

func TestHookRegistry(t *testing.T) {
	hooks := framework.NewHookRegistry()
	called := false
	hooks.Register(framework.EventBeforeTool, func(_ context.Context, _ framework.HookContext) error {
		called = true
		return nil
	})
	if err := hooks.Emit(context.Background(), framework.EventBeforeTool, framework.HookContext{ToolName: "x"}); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("hook not called")
	}
}
