package contextbuilder

import (
	"context"
	"strings"
)

func (b *Builder) history(ctx context.Context, sessionID string) string {
	if sessionID == "" || b.db == nil {
		return ""
	}
	limit := b.cfg.Agent.MaxHistoryMessages
	if limit <= 0 {
		limit = 12
	}
	rows, err := b.db.QueryContext(ctx, `
SELECT role, content FROM messages
WHERE session_id = ?
ORDER BY id DESC
LIMIT ?`, sessionID, limit)
	if err != nil {
		return ""
	}
	defer rows.Close()

	var lines []string
	for rows.Next() {
		var role, content string
		if rows.Scan(&role, &content) == nil {
			lines = append(lines, role+": "+content)
		}
	}
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}
	return trimTo(strings.Join(lines, "\n"), b.cfg.Agent.MaxContextChars/2)
}
