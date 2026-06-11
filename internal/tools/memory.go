package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"vietclaw/internal/memory"
	"vietclaw/internal/providers"
)

const (
	toolMemoryRecall = "memory_recall"
	toolMemoryStore  = "memory_store"
)

type MemoryRecall struct {
	Store *memory.Store
}

func (t MemoryRecall) Name() string { return toolMemoryRecall }

func (t MemoryRecall) Run(ctx context.Context, argsJSON string) (string, error) {
	if t.Store == nil {
		return "", fmt.Errorf("memory store not configured")
	}
	var args struct {
		Query string `json:"query"`
		Kind  string `json:"kind"`
		Limit int    `json:"limit"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", err
	}
	query := strings.TrimSpace(args.Query)
	if query == "" {
		return "", fmt.Errorf("query is required")
	}
	limit := args.Limit
	if limit <= 0 || limit > 20 {
		limit = 5
	}
	scope := MemoryScopeFrom(ctx)
	records, err := t.Store.SearchHybrid(ctx, scope, query, limit, EmbedderFrom(ctx))
	if err != nil {
		return "", err
	}
	if args.Kind != "" {
		filtered := make([]memory.Record, 0, len(records))
		for _, rec := range records {
			if string(rec.Kind) == args.Kind {
				filtered = append(filtered, rec)
			}
		}
		records = filtered
	}
	if len(records) == 0 {
		return "No matching memories found.", nil
	}
	lines := make([]string, 0, len(records))
	for _, rec := range records {
		lines = append(lines, fmt.Sprintf("[%s] %s", rec.Kind, rec.Content))
	}
	return strings.Join(lines, "\n"), nil
}

type MemoryStore struct {
	Store *memory.Store
}

func (t MemoryStore) Name() string { return toolMemoryStore }

func (t MemoryStore) Run(ctx context.Context, argsJSON string) (string, error) {
	if t.Store == nil {
		return "", fmt.Errorf("memory store not configured")
	}
	var args struct {
		Content    string `json:"content"`
		Kind       string `json:"kind"`
		Confidence string `json:"confidence"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", err
	}
	content := strings.TrimSpace(args.Content)
	if content == "" {
		return "", fmt.Errorf("content is required")
	}
	kind := memory.KindNote
	if args.Kind != "" {
		kind = memory.Kind(args.Kind)
	}
	confidence := memory.ConfidenceConfirmed
	switch strings.ToLower(strings.TrimSpace(args.Confidence)) {
	case "inferred":
		confidence = memory.ConfidenceInferred
	case "temporary":
		confidence = memory.ConfidenceTemporary
	}

	var embedding []float32
	if embedder := EmbedderFrom(ctx); embedder != nil {
		embedding, _ = embedder.Embed(ctx, content)
	}

	rec, err := t.Store.Add(ctx, memory.Record{
		Scope:      MemoryScopeFrom(ctx),
		Kind:       kind,
		Content:    content,
		Confidence: confidence,
		Embedding:  embedding,
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Stored memory #%d: %s", rec.ID, rec.Content), nil
}

func memoryToolDefinitions(lang string) []providers.ToolDefinition {
	return []providers.ToolDefinition{
		{
			Type: "function",
			Function: providers.FunctionDetail{
				Name:        toolMemoryRecall,
				Description: "Search long-term memory for facts, preferences, or past lessons. Use when context is missing or before complex tasks (MemGPT-style active recall).",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"query": map[string]any{
							"type":        "string",
							"description": "Natural-language search query.",
						},
						"kind": map[string]any{
							"type":        "string",
							"description": "Optional filter: profile, preference, project, workflow, decision, experience, note.",
						},
						"limit": map[string]any{
							"type":        "integer",
							"description": "Max results (default 5).",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		{
			Type: "function",
			Function: providers.FunctionDetail{
				Name:        toolMemoryStore,
				Description: "Persist a durable fact, preference, or lesson into long-term memory for future sessions.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"content": map[string]any{
							"type":        "string",
							"description": "Memory text to store.",
						},
						"kind": map[string]any{
							"type":        "string",
							"description": "Memory category: profile, preference, project, workflow, decision, experience, note.",
						},
						"confidence": map[string]any{
							"type":        "string",
							"description": "confirmed, inferred, or temporary.",
						},
					},
					"required": []string{"content"},
				},
			},
		},
	}
}
