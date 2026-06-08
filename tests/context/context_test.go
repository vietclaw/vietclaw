package context_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"vietclaw/internal/config"
	contextbuilder "vietclaw/internal/context"
	"vietclaw/internal/db"
	"vietclaw/internal/memory"
)

func TestBuilderLimitsContextChars(t *testing.T) {
	database, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := db.ApplySchema(database); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default(config.Paths{DataDir: t.TempDir()})
	cfg.Agent.MaxContextChars = 120
	store := memory.NewStore(database)
	if _, err := store.Add(context.Background(), memory.Record{
		Scope:   "user:local",
		Kind:    memory.KindNote,
		Content: strings.Repeat("token ", 200),
	}); err != nil {
		t.Fatal(err)
	}

	builder := contextbuilder.New(cfg, database, store)
	messages, err := builder.Messages(context.Background(), "", "local", "token", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len([]rune(messages[0].Content)) > 120 {
		t.Fatalf("context len = %d", len([]rune(messages[0].Content)))
	}
}

func TestBuilderInjectsMatchingSkill(t *testing.T) {
	database, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := db.ApplySchema(database); err != nil {
		t.Fatal(err)
	}

	dir := filepath.Join(t.TempDir(), "skills", "review")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# Code Review\ntriggers: review\nAlways inspect risks first."), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default(config.Paths{DataDir: t.TempDir()})
	cfg.Agent.SkillDirs = []string{filepath.Dir(dir)}
	store := memory.NewStore(database)
	builder := contextbuilder.New(cfg, database, store)
	messages, err := builder.Messages(context.Background(), "", "local", "please review this", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(messages[0].Content, "Always inspect risks first.") {
		t.Fatalf("skill instructions not injected: %q", messages[0].Content)
	}
}
