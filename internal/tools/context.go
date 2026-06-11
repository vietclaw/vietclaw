package tools

import (
	"context"

	"vietclaw/internal/providers"
)

type ctxKey string

const (
	ctxMemoryScope ctxKey = "memory_scope"
	ctxEmbedder    ctxKey = "embedder"
)

func WithMemoryScope(ctx context.Context, scope string) context.Context {
	return context.WithValue(ctx, ctxMemoryScope, scope)
}

func MemoryScopeFrom(ctx context.Context) string {
	v, _ := ctx.Value(ctxMemoryScope).(string)
	return v
}

func WithEmbedder(ctx context.Context, embedder providers.Provider) context.Context {
	return context.WithValue(ctx, ctxEmbedder, embedder)
}

func EmbedderFrom(ctx context.Context) providers.Provider {
	v, _ := ctx.Value(ctxEmbedder).(providers.Provider)
	return v
}
