package agentfs_test

import (
	"os"
	"path/filepath"
	"testing"

	"vietclaw/internal/agentfs"
	"vietclaw/internal/config"
)

func TestParseAgentAndSkills(t *testing.T) {
	root := t.TempDir()
	agentPath := filepath.Join(root, "researcher", agentfs.AgentFileName)
	if err := agentfs.WriteAgent(agentPath, agentfs.CreateRequest{
		ID:       "researcher",
		Name:     "Researcher",
		Language: "vi",
		Persona:  "Focus on research.",
		Tools:    []string{"web_search"},
		Spawnable: true,
		Skills: []agentfs.SkillInput{{
			Name:         "review",
			Triggers:     []string{"review"},
			Instructions: "Inspect risks first.",
		}},
		ToolGuides: []agentfs.ToolGuideInput{{
			Tool:         "web_search",
			Instructions: "Prefer Vietnamese sources.",
		}},
	}); err != nil {
		t.Fatal(err)
	}

	registry := agentfs.NewRegistry(root, config.Default(config.Paths{DataDir: root}))
	if err := registry.Reload(); err != nil {
		t.Fatal(err)
	}
	def, ok := registry.Get("researcher")
	if !ok {
		t.Fatal("researcher not found")
	}
	if def.Persona != "Focus on research." {
		t.Fatalf("persona = %q", def.Persona)
	}
	if len(def.Skills) != 1 || def.Skills[0].Name != "review" {
		t.Fatalf("skills = %#v", def.Skills)
	}
	if len(def.ToolGuides) != 1 || def.ToolGuides[0].Tool != "web_search" {
		t.Fatalf("tool guides = %#v", def.ToolGuides)
	}
}

func TestMigrateFromConfig(t *testing.T) {
	root := t.TempDir()
	cfg := config.Default(config.Paths{DataDir: root})
	cfg.Agents = []config.AgentProfileConfig{{
		ID:      "coder",
		Name:    "Coder",
		Persona: "Write code.",
	}}
	migrated, err := agentfs.MigrateFromConfig(nil, config.Paths{DataDir: root}, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !migrated {
		t.Fatal("expected migration")
	}
	if _, err := os.Stat(filepath.Join(root, "agents", "coder", agentfs.AgentFileName)); err != nil {
		t.Fatal(err)
	}
	if len(cfg.Agents) != 0 {
		t.Fatalf("agents should be cleared: %#v", cfg.Agents)
	}
}
