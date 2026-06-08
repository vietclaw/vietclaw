package contextbuilder

import (
	"context"
	"database/sql"
	"strings"

	"vietclaw/internal/config"
	"vietclaw/internal/i18n"
	"vietclaw/internal/memory"
	"vietclaw/internal/providers"
	"vietclaw/internal/router"
	"vietclaw/internal/skills"
)

type Builder struct {
	cfg    config.Config
	db     *sql.DB
	mem    *memory.Store
	router *router.ModelRouter
}

func New(cfg config.Config, db *sql.DB, mem *memory.Store) *Builder {
	return &Builder{cfg: cfg, db: db, mem: mem}
}

func (b *Builder) WithRouter(r *router.ModelRouter) *Builder {
	b.router = r
	return b
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

	if sessionID != "" && b.router != nil {
		go b.summarizeIfNeeded(context.Background(), sessionID, b.router)
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

	prompt := "Tóm tắt ngắn gọn các ý chính và diễn biến cuộc hội thoại dưới đây bằng tiếng Việt để làm ngữ cảnh cho AI. Chỉ trả về văn bản tóm tắt ngắn gọn, không có thêm lời dẫn hay định dạng khác:\n\n" + strings.Join(conv, "\n")
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
