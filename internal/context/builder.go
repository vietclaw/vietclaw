package contextbuilder

import (
	"context"
	"database/sql"
	"strings"

	"vietclaw/internal/config"
	"vietclaw/internal/i18n"
	"vietclaw/internal/memory"
	"vietclaw/internal/providers"
	"vietclaw/internal/skills"
)

type Builder struct {
	cfg config.Config
	db  *sql.DB
	mem *memory.Store
}

func New(cfg config.Config, db *sql.DB, mem *memory.Store) *Builder {
	return &Builder{cfg: cfg, db: db, mem: mem}
}

func (b *Builder) Messages(ctx context.Context, sessionID, userID, userMessage string, embedder providers.Provider) ([]providers.Message, error) {
	maxChars := b.cfg.Agent.MaxContextChars
	if maxChars <= 0 {
		maxChars = config.DefaultMaxContextChars
	}

	lang := b.cfg.Agent.Language
	agentName := b.cfg.Agent.Name
	if agentName == "" {
		agentName = config.AppName
	}

	parts := []string{i18n.T(lang, i18n.SystemPromptBase, agentName)}
	scope := scopeForUser(userID)
	memories, _ := b.mem.SearchHybrid(ctx, scope, userMessage, 6, embedder)
	if len(memories) > 0 {
		lines := []string{i18n.T(lang, i18n.SystemMemoryHeader)}
		for _, rec := range memories {
			lines = append(lines, "- "+rec.Content)
		}
		parts = append(parts, strings.Join(lines, "\n"))
	}

	loadedSkills, _ := skills.Load(b.cfg.Agent.SkillDirs)
	matchedSkills := skills.Match(loadedSkills, userMessage, 3)
	if len(matchedSkills) > 0 {
		lines := []string{i18n.T(lang, i18n.SystemSkillHeader)}
		for _, skill := range matchedSkills {
			lines = append(lines, "- "+skill.Name+": "+skill.Instructions)
		}
		parts = append(parts, strings.Join(lines, "\n"))
	}

	history := b.history(ctx, sessionID)
	if history != "" {
		parts = append(parts, i18n.T(lang, i18n.SystemHistoryHeader)+"\n"+history)
	}

	system := trimTo(strings.Join(parts, "\n\n"), maxChars)
	return []providers.Message{
		{Role: "system", Content: system},
		{Role: "user", Content: userMessage},
	}, nil
}

func scopeForUser(userID string) string {
	if strings.Contains(userID, ":") {
		return userID
	}
	return "user:" + userID
}
