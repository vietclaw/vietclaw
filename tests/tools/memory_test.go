package tools_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"vietclaw/internal/config"
	"vietclaw/internal/db"
	"vietclaw/internal/memory"
	"vietclaw/internal/tools"
)

func TestMemoryStoreAndRecall(t *testing.T) {
	database, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := db.ApplySchema(database); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default(config.Paths{DataDir: t.TempDir()})
	registry := tools.NewRegistry(cfg).WithMemory(memory.NewStore(database))
	ctx := tools.WithMemoryScope(context.Background(), "user:test")

	storeOut, err := registry.Execute(ctx, "memory_store", `{"content":"prefers dark mode","kind":"preference"}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(storeOut, "prefers dark mode") {
		t.Fatalf("store output: %s", storeOut)
	}

	recallOut, err := registry.Execute(ctx, "memory_recall", `{"query":"dark mode"}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(recallOut, "prefers dark mode") {
		t.Fatalf("recall output: %s", recallOut)
	}
}
