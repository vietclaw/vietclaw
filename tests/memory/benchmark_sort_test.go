package memory_test

import (
	"math/rand"
	"sort"
	"testing"
	"vietclaw/internal/memory"
)

func BenchmarkSortOriginal(b *testing.B) {
	// Create test data
	var records []memory.Record
	for i := 0; i < 1000; i++ {
		emb := make([]float32, 128)
		for j := range emb {
			emb[j] = rand.Float32()
		}
		records = append(records, memory.Record{Embedding: emb})
	}

	queryEmb := make([]float32, 128)
	for j := range queryEmb {
		queryEmb[j] = rand.Float32()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// clone the records so we're always sorting unsorted/randomly ordered data
		testRecords := make([]memory.Record, len(records))
		copy(testRecords, records)
		b.StartTimer()

		sort.Slice(testRecords, func(i, j int) bool {
			return memory.CosineSimilarity(queryEmb, testRecords[i].Embedding) > memory.CosineSimilarity(queryEmb, testRecords[j].Embedding)
		})
	}
}

func BenchmarkSortOptimized(b *testing.B) {
	// Create test data
	var records []memory.Record
	for i := 0; i < 1000; i++ {
		emb := make([]float32, 128)
		for j := range emb {
			emb[j] = rand.Float32()
		}
		records = append(records, memory.Record{Embedding: emb})
	}

	queryEmb := make([]float32, 128)
	for j := range queryEmb {
		queryEmb[j] = rand.Float32()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// clone the records
		testRecords := make([]memory.Record, len(records))
		copy(testRecords, records)
		b.StartTimer()

		type scoredRecord struct {
			record memory.Record
			score  float32
		}

		scored := make([]scoredRecord, len(testRecords))
		for k, rec := range testRecords {
			scored[k] = scoredRecord{
				record: rec,
				score:  memory.CosineSimilarity(queryEmb, rec.Embedding),
			}
		}

		sort.Slice(scored, func(i, j int) bool {
			return scored[i].score > scored[j].score
		})

		for k := range scored {
			testRecords[k] = scored[k].record
		}
	}
}
