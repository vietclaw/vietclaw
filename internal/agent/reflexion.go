package agent

import (
	"context"
	"fmt"
	"strings"

	"vietclaw/internal/config"
	"vietclaw/internal/memory"
	"vietclaw/internal/providers"
	"vietclaw/internal/tools"
)

// recordToolReflection stores a Reflexion-style verbal lesson when a tool fails.
// See: Shinn et al., "Reflexion: Language Agents with Verbal Reinforcement Learning" (2023).
func (s *Service) recordToolReflection(ctx context.Context, scope string, toolName, argsJSON, errText string, embedder providers.Provider) {
	if !s.cfg.Agent.Reflexion.Enabled || s.mem == nil {
		return
	}
	errText = strings.TrimSpace(errText)
	if errText == "" {
		return
	}
	argsJSON = strings.TrimSpace(argsJSON)
	if len(argsJSON) > 200 {
		argsJSON = argsJSON[:200] + "..."
	}
	lesson := fmt.Sprintf("Tool %s failed: %s. Input: %s", toolName, errText, argsJSON)

	var embedding []float32
	if embedder != nil {
		embedding, _ = embedder.Embed(ctx, lesson)
	}
	_, _ = s.mem.Add(ctx, memory.Record{
		Scope:      scope,
		Kind:       memory.KindExperience,
		Content:    lesson,
		Confidence: memory.ConfidenceConfirmed,
		Embedding:  embedding,
	})
}

func (s *Service) maxAgentSteps() int {
	if s.cfg.Agent.MaxSteps > 0 {
		return s.cfg.Agent.MaxSteps
	}
	if s.cfg.Agent.MaxSteps == 0 {
		return 0
	}
	return config.DefaultMaxAgentSteps
}

func (s *Service) toolContext(ctx context.Context, req ChatRequest) context.Context {
	ctx = tools.WithMemoryScope(ctx, s.memoryScope(req))
	if embedder := s.router.SelectDefaultEmbedder(); embedder != nil {
		ctx = tools.WithEmbedder(ctx, embedder)
	}
	return ctx
}
