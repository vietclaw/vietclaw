package agent_test

import (
	"testing"

	"vietclaw/internal/config"
)

func TestDefaultMaxSteps(t *testing.T) {
	cfg := config.Default(config.Paths{DataDir: t.TempDir()})
	if cfg.Agent.MaxSteps != config.DefaultMaxAgentSteps {
		t.Fatalf("max steps = %d want %d", cfg.Agent.MaxSteps, config.DefaultMaxAgentSteps)
	}
	if !cfg.Agent.Reflexion.Enabled {
		t.Fatal("reflexion should be enabled by default")
	}
	if !cfg.Agent.MemoryTools.Enabled {
		t.Fatal("memory tools should be enabled by default")
	}
}
