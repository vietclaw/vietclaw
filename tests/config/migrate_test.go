package config_test

import (
	"testing"

	"vietclaw/internal/config"
)

func TestApplyLegacyMigrationsMaxSteps(t *testing.T) {
	cfg := config.Config{
		Agent: config.AgentConfig{MaxSteps: 5, MaxOutputTokens: 512},
		Agents: []config.AgentProfileConfig{
			{ID: "default", MaxSteps: 12},
		},
	}
	got := config.ApplyLegacyMigrations(cfg)
	if got.Agent.MaxSteps != 0 {
		t.Fatalf("agent max_steps = %d want 0", got.Agent.MaxSteps)
	}
	if got.Agent.MaxOutputTokens != 0 {
		t.Fatalf("max_output_tokens = %d want 0 (unlimited)", got.Agent.MaxOutputTokens)
	}
	if got.Agents[0].MaxSteps != 0 {
		t.Fatalf("profile max_steps = %d want 0", got.Agents[0].MaxSteps)
	}
}

func TestApplyLegacyMigrationsPreservesCustomCap(t *testing.T) {
	cfg := config.Config{Agent: config.AgentConfig{MaxSteps: 20, MaxOutputTokens: 0}}
	got := config.ApplyLegacyMigrations(cfg)
	if got.Agent.MaxSteps != 20 {
		t.Fatalf("custom max_steps changed to %d", got.Agent.MaxSteps)
	}
	if got.Agent.MaxOutputTokens != 0 {
		t.Fatalf("max_output_tokens changed to %d", got.Agent.MaxOutputTokens)
	}
}
