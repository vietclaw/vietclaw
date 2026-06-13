package memory_test

import (
	"context"
	"path/filepath"
	"testing"

	"vietclaw/internal/db"
	"vietclaw/internal/memory"
)

func TestStoreDeleteMany(t *testing.T) {
	database, err := db.Open(filepath.Join(t.TempDir(), "test_deletemany.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := db.ApplySchema(database); err != nil {
		t.Fatal(err)
	}

	store := memory.NewStore(database)
	ctx := context.Background()

	var ids []int64
	for i := 0; i < 10; i++ {
		rec, err := store.Add(ctx, memory.Record{
			Scope:   "test",
			Content: "content",
		})
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, rec.ID)
	}

	recs, err := store.List(ctx, "test", 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 10 {
		t.Fatalf("expected 10 records, got %d", len(recs))
	}

	err = store.DeleteMany(ctx, ids[:5])
	if err != nil {
		t.Fatal(err)
	}

	recs, err = store.List(ctx, "test", 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 5 {
		t.Fatalf("expected 5 records left, got %d", len(recs))
	}

	// Ensure we can delete empty
	err = store.DeleteMany(ctx, []int64{})
	if err != nil {
		t.Fatal(err)
	}
}
