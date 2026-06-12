package contextbuilder

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"golang.org/x/sync/singleflight"

	"vietclaw/internal/agentfs"
	"vietclaw/internal/config"
	"vietclaw/internal/i18n"
	"vietclaw/internal/memory"
	"vietclaw/internal/providers"
	"vietclaw/internal/router"
	"vietclaw/internal/skills"
)

type Builder struct {
	cfg            config.Config
	db             *sql.DB
	mem            *memory.Store
	router         *router.ModelRouter
	agents         *agentfs.Registry
	summarizeGroup singleflight.Group
}

func New(cfg config.Config, db *sql.DB, mem *memory.Store) *Builder {
	return &Builder{cfg: cfg, db: db, mem: mem}
}

func (b *Builder) WithRouter(r *router.ModelRouter) *Builder {
	b.router = r
	return b
}

func (b *Builder) WithAgentRegistry(registry *agentfs.Registry) *Builder {
	b.agents = registry
	return b
}

func (b *Builder) Messages(ctx context.Context, sessionID, scope, agentID, userMessage string, embedder providers.Provider) ([]providers.Message, error) {
	maxChars := b.cfg.Agent.MaxContextChars
	if maxChars <= 0 {
		maxChars = config.DefaultMaxContextChars
	}

	lang := b.cfg.Agent.Language
	agentName := b.cfg.Agent.Name
	if agentName == "" {
		agentName = config.AppName
	}

	if agentID != "" && b.agents != nil {
		if def, ok := b.agents.Get(agentID); ok {
			if strings.TrimSpace(def.Name) != "" {
				agentName = def.Name
			}
			if strings.TrimSpace(def.Language) != "" {
				lang = def.Language
			}
		}
	}

	parts := []string{i18n.T(lang, i18n.SystemPromptBase, agentName)}
	if scope == "" {
		scope = scopeForUser("local")
	}
	hybrid, _ := b.mem.SearchHybrid(ctx, scope, userMessage, 9, embedder)
	coreMemories, experiences := splitMemories(hybrid, 6, 3)
	if len(coreMemories) > 0 || len(experiences) > 0 {
		lines := []string{i18n.T(lang, i18n.SystemMemoryHeader)}
		for _, rec := range coreMemories {
			lines = append(lines, "- "+rec.Content)
		}
		for _, rec := range experiences {
			lines = append(lines, "- [lesson] "+rec.Content)
		}
		parts = append(parts, strings.Join(lines, "\n"))
	}

	var loadedSkills []skills.Skill
	if b.agents != nil {
		loadedSkills = b.agents.SkillsFor(agentID)
	}
	if len(loadedSkills) == 0 {
		loadedSkills, _ = skills.Load(b.cfg.Agent.SkillDirs)
	}
	matchedSkills := skills.Match(loadedSkills, userMessage, 3)
	if len(matchedSkills) > 0 {
		lines := []string{i18n.T(lang, i18n.SystemSkillHeader)}
		for _, skill := range matchedSkills {
			lines = append(lines, "- "+skill.Name+": "+skill.Instructions)
		}
		parts = append(parts, strings.Join(lines, "\n"))
	}

	if agentBlock := b.agentContextBlock(lang); agentBlock != "" {
		parts = append(parts, agentBlock)
	}

	history := b.history(ctx, sessionID)
	if history != "" {
		parts = append(parts, i18n.T(lang, i18n.SystemHistoryHeader)+"\n"+history)
	}

	if sessionID != "" && b.router != nil {
		go func() {
			_, _, _ = b.summarizeGroup.Do(sessionID, func() (any, error) {
				b.summarizeIfNeeded(context.Background(), sessionID, b.router)
				return nil, nil
			})
		}()
	}

	system := trimTo(strings.Join(parts, "\n\n"), maxChars)
	return []providers.Message{
		{Role: "system", Content: system},
		{Role: "user", Content: userMessage},
	}, nil
}

func splitMemories(records []memory.Record, coreLimit, expLimit int) ([]memory.Record, []memory.Record) {
	core := make([]memory.Record, 0, coreLimit)
	exp := make([]memory.Record, 0, expLimit)
	for _, rec := range records {
		if rec.Kind == memory.KindExperience {
			if len(exp) < expLimit {
				exp = append(exp, rec)
			}
			continue
		}
		if len(core) < coreLimit {
			core = append(core, rec)
		}
	}
	return core, exp
}

func scopeForUser(userID string) string {
	if strings.Contains(userID, ":") {
		return userID
	}
	return "user:" + userID
}

func (b *Builder) summarizeIfNeeded(ctx context.Context, sessionID string, r *router.ModelRouter) {
	if sessionID == "" || b.db == nil || r == nil {
		return
	}
	limit := b.cfg.Agent.MaxHistoryMessages
	if limit <= 0 {
		limit = 12
	}

	var count int
	err := b.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM messages WHERE session_id = ?", sessionID).Scan(&count)
	if err != nil || count <= limit {
		return
	}

	// Summarize every 6 messages after the limit
	if (count-limit)%6 != 0 {
		return
	}

	rows, err := b.db.QueryContext(ctx, "SELECT role, content FROM messages WHERE session_id = ? ORDER BY id ASC", sessionID)
	if err != nil {
		return
	}
	defer rows.Close()

	var conv []string
	for rows.Next() {
		var role, content string
		if rows.Scan(&role, &content) == nil {
			conv = append(conv, role+": "+content)
		}
	}
	if len(conv) == 0 {
		return
	}

	selection, err := r.Select(ctx, providers.ChatRequest{
		Messages: []providers.Message{{Role: "user", Content: "dummy"}},
	}, nil)
	if err != nil || selection.Provider == nil {
		return
	}

	prompt := fmt.Sprintf(i18n.T(b.cfg.Agent.Language, i18n.SystemSummarizePrompt), strings.Join(conv, "\n"))
	resp, err := selection.Provider.Chat(ctx, providers.ChatRequest{
		Model:           selection.Model,
		MaxOutputTokens: 256,
		Temperature:     0.3,
		Messages: []providers.Message{
			{Role: "user", Content: prompt},
		},
	})
	if err != nil || resp.Text == "" {
		return
	}

	_, _ = b.db.ExecContext(ctx, "UPDATE sessions SET summary = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", strings.TrimSpace(resp.Text), sessionID)
}
