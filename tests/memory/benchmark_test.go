package memory_test

import (
	"context"
	"math/rand"
	"path/filepath"
	"testing"

	"vietclaw/internal/db"
	"vietclaw/internal/memory"
	"vietclaw/internal/providers"
)

type mockEmbedderBench struct {
	embedding []float32
}

func (m mockEmbedderBench) ID() string { return "mock" }
func (m mockEmbedderBench) Type() string { return "mock" }
func (m mockEmbedderBench) Chat(ctx context.Context, req providers.ChatRequest) (providers.ChatResponse, error) {
	return providers.ChatResponse{}, nil
}
func (m mockEmbedderBench) ChatStream(ctx context.Context, req providers.ChatRequest) (<-chan providers.StreamChunk, error) {
	return nil, nil
}
func (m mockEmbedderBench) EstimateCost(req providers.ChatRequest) providers.CostEstimate {
	return providers.CostEstimate{}
}
func (m mockEmbedderBench) Embed(ctx context.Context, text string) ([]float32, error) {
	return m.embedding, nil
}

func BenchmarkSearchVectorCandidates(b *testing.B) {
	database, err := db.Open(filepath.Join(b.TempDir(), "bench.db"))
	if err != nil {
		b.Fatal(err)
	}
	defer database.Close()
	if err := db.ApplySchema(database); err != nil {
		b.Fatal(err)
	}

	store := memory.NewStore(database)

	// Insert dummy records
	for i := 0; i < 1000; i++ {
		emb := make([]float32, 128)
		for j := range emb {
			emb[j] = rand.Float32()
		}

		_, err := store.Add(context.Background(), memory.Record{
			Scope:      "bench",
			Kind:       memory.KindNote,
			Content:    "bench content",
			Confidence: memory.ConfidenceConfirmed,
			Embedding:  emb,
		})
		if err != nil {
			b.Fatal(err)
		}
	}

	queryEmb := make([]float32, 128)
	for j := range queryEmb {
		queryEmb[j] = rand.Float32()
	}

	embedder := mockEmbedderBench{embedding: queryEmb}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := store.SearchHybrid(context.Background(), "bench", "test query", 100, embedder)
		if err != nil {
			b.Fatal(err)
		}
	}
}
