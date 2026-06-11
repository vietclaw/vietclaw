package config_test

import (
	"testing"

	"vietclaw/internal/config"
)

func TestJsonPathExists(t *testing.T) {
	raw := []byte(`{"agent":{"max_steps":0,"heartbeat":{"enabled":false}}}`)
	if !config.JsonPathExists(raw, "agent", "max_steps") {
		t.Fatal("expected max_steps path")
	}
	if config.JsonPathExists(raw, "agent", "reflexion") {
		t.Fatal("reflexion should be missing")
	}
}

func TestMergeAgentOptionalRespectsExplicitZeroMaxSteps(t *testing.T) {
	raw := []byte(`{"agent":{"max_steps":0}}`)
	def := config.Default(config.Paths{DataDir: t.TempDir()})
	cfg := config.AgentConfig{MaxSteps: 0}
	got := config.MergeAgentOptional(cfg, def.Agent, raw)
	if got.MaxSteps != 0 {
		t.Fatalf("max_steps = %d want 0 (unlimited)", got.MaxSteps)
	}
}

func TestMergeAgentOptionalAppliesDefaultsWhenMissing(t *testing.T) {
	raw := []byte(`{"agent":{"name":"VietClaw"}}`)
	def := config.Default(config.Paths{DataDir: t.TempDir()})
	cfg := config.AgentConfig{Name: "VietClaw"}
	got := config.MergeAgentOptional(cfg, def.Agent, raw)
	if got.MaxSteps != def.Agent.MaxSteps {
		t.Fatalf("max_steps = %d want %d", got.MaxSteps, def.Agent.MaxSteps)
	}
	if !got.Reflexion.Enabled {
		t.Fatal("reflexion should default enabled")
	}
	if !got.MemoryTools.Enabled {
		t.Fatal("memory tools should default enabled")
	}
	if got.Experience != def.Agent.Experience {
		t.Fatalf("experience = %q want %q", got.Experience, def.Agent.Experience)
	}
}
