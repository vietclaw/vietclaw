package memory_test

import (
	"context"
	"path/filepath"
	"math"
	"testing"

	"vietclaw/internal/db"
	"vietclaw/internal/memory"
)

func TestVectorConvert(t *testing.T) {
	tests := []struct {
		name  string
		slice []float32
	}{
		{"regular values", []float32{1.0, -2.5, 3.14, 0.0}},
		{"empty slice", []float32{}},
		{"nil slice", nil},
		{"special float values", []float32{float32(math.NaN()), float32(math.Inf(1)), float32(math.Inf(-1))}},
		{"large slice", make([]float32, 1000)}, // Large slice of 0s
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := memory.Float32SliceToBytes(tc.slice)
			restored := memory.BytesToFloat32Slice(data)

			if len(tc.slice) == 0 {
				if len(restored) != 0 {
					t.Fatalf("expected empty/nil slice, got len %d", len(restored))
				}
				return
			}

			if len(restored) != len(tc.slice) {
				t.Fatalf("expected len %d, got %d", len(tc.slice), len(restored))
			}

			for i := range tc.slice {
				if math.IsNaN(float64(tc.slice[i])) {
					if !math.IsNaN(float64(restored[i])) {
						t.Errorf("expected NaN at index %d, got %f", i, restored[i])
					}
				} else if tc.slice[i] != restored[i] {
					t.Errorf("expected %f at index %d, got %f", tc.slice[i], i, restored[i])
				}
			}
		})
	}
}

func TestBytesToFloat32Slice_InvalidLength(t *testing.T) {
	tests := [][]byte{
		{1, 2, 3},       // length 3
		{1, 2, 3, 4, 5}, // length 5
	}

	for _, buf := range tests {
		result := memory.BytesToFloat32Slice(buf)
		if result != nil {
			t.Errorf("expected nil for buffer of length %d, got slice of length %d", len(buf), len(result))
		}
	}
}

func TestCosineSimilarity(t *testing.T) {
	v1 := []float32{1.0, 0.0, 0.0}
	v2 := []float32{1.0, 0.0, 0.0}
	if sim := memory.CosineSimilarity(v1, v2); sim < 0.99 {
		t.Errorf("expected identity to be ~1.0, got %f", sim)
	}

	v3 := []float32{0.0, 1.0, 0.0}
	if sim := memory.CosineSimilarity(v1, v3); sim > 0.01 {
		t.Errorf("expected orthogonal vectors similarity to be ~0.0, got %f", sim)
	}
}

func TestHybridSearchFindsVectorOnlyCandidate(t *testing.T) {
	database, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := db.ApplySchema(database); err != nil {
		t.Fatal(err)
	}

	store := memory.NewStore(database)
	if _, err := store.Add(context.Background(), memory.Record{
		Scope:      "user:local",
		Kind:       memory.KindNote,
		Content:    "deploy runbook lives in ops notebook",
		Confidence: memory.ConfidenceConfirmed,
		Embedding:  []float32{1, 0, 0},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Add(context.Background(), memory.Record{
		Scope:      "user:local",
		Kind:       memory.KindNote,
		Content:    "favorite color is blue",
		Confidence: memory.ConfidenceConfirmed,
		Embedding:  []float32{0, 1, 0},
	}); err != nil {
		t.Fatal(err)
	}

	results, err := store.SearchHybrid(context.Background(), "user:local", "release procedure", 1, fixedEmbedder{embedding: []float32{1, 0, 0}})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Content != "deploy runbook lives in ops notebook" {
		t.Fatalf("hybrid vector results = %#v", results)
	}
}
