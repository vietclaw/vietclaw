package memory_test

import (
	"context"
	"path/filepath"
	"testing"

	"vietclaw/internal/db"
	"vietclaw/internal/memory"
)

func TestStoreAddAndSearch(t *testing.T) {
	database, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := db.ApplySchema(database); err != nil {
		t.Fatal(err)
	}

	store := memory.NewStore(database)
	rec, err := store.Add(context.Background(), memory.Record{
		Scope:      "user:local",
		Kind:       memory.KindNote,
		Content:    "Minh thích tiết kiệm token",
		Confidence: memory.ConfidenceConfirmed,
	})
	if err != nil {
		t.Fatal(err)
	}
	if rec.ID == 0 {
		t.Fatal("expected persisted id")
	}

	results, err := store.Search(context.Background(), "user:local", "token", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Content != "Minh thích tiết kiệm token" {
		t.Fatalf("search results = %#v", results)
	}
}
