package context_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"vietclaw/internal/agentfs"
	"vietclaw/internal/config"
	contextbuilder "vietclaw/internal/context"
	"vietclaw/internal/db"
	"vietclaw/internal/memory"
)

func TestBuilderInjectsAgentContext(t *testing.T) {
	database, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := db.ApplySchema(database); err != nil {
		t.Fatal(err)
	}

	dataDir := t.TempDir()
	cfg := config.Default(config.Paths{DataDir: dataDir})
	cfg.Framework.Enabled = true
	cfg.Framework.AllowAutoCreate = false

	registry := agentfs.NewRegistry(agentfs.DefaultRoot(dataDir), cfg)
	if err := registry.Reload(); err != nil {
		t.Fatal(err)
	}

	store := memory.NewStore(database)
	builder := contextbuilder.New(cfg, database, store).WithAgentRegistry(registry)
	messages, err := builder.Messages(context.Background(), "", "user:local", config.DefaultAgentID, "spawn a reviewer", nil)
	if err != nil {
		t.Fatal(err)
	}
	content := messages[0].Content
	if !strings.Contains(content, "agent_create") || !strings.Contains(content, "KHÔNG") {
		t.Fatalf("expected auto-create disabled hint, got: %q", content)
	}
}

func TestBuilderInjectsAgentCreateGuideWhenEnabled(t *testing.T) {
	database, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := db.ApplySchema(database); err != nil {
		t.Fatal(err)
	}

	dataDir := t.TempDir()
	cfg := config.Default(config.Paths{DataDir: dataDir})
	cfg.Framework.Enabled = true
	cfg.Framework.AllowAutoCreate = true

	registry := agentfs.NewRegistry(agentfs.DefaultRoot(dataDir), cfg)
	if err := registry.Reload(); err != nil {
		t.Fatal(err)
	}

	store := memory.NewStore(database)
	builder := contextbuilder.New(cfg, database, store).WithAgentRegistry(registry)
	messages, err := builder.Messages(context.Background(), "", "user:local", config.DefaultAgentID, "create a reviewer agent", nil)
	if err != nil {
		t.Fatal(err)
	}
	content := messages[0].Content
	if !strings.Contains(content, "agent_create") || !strings.Contains(content, "AGENT.md") {
		t.Fatalf("expected auto-create guide, got: %q", content)
	}
}
