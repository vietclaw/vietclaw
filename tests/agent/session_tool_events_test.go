package agent_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"vietclaw/internal/agent"
	"vietclaw/internal/config"
	"vietclaw/internal/db"
)

func TestSessionMessagesIncludesToolEvents(t *testing.T) {
	database, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := db.ApplySchema(database); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default(config.Paths{DataDir: t.TempDir()})
	svc := agent.NewService(cfg, database)
	ctx := context.Background()

	sessionID := "sess_parent:spawn:researcher:sub_abc:delegate:researcher"
	now := time.Now().UTC().Format(time.RFC3339)

	if _, err := database.ExecContext(ctx, `
INSERT INTO sessions (id, channel, user_id, title, summary, created_at, updated_at)
VALUES (?, 'web', 'local', NULL, NULL, ?, ?)`, sessionID, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `
INSERT INTO messages (session_id, role, content, created_at)
VALUES (?, 'user', 'Compare Redis and SQLite', ?)`, sessionID, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `
INSERT INTO messages (session_id, role, content, created_at)
VALUES (?, 'assistant', 'Here is the comparison', ?)`, sessionID, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `
INSERT INTO agent_runs (id, session_id, parent_run_id, intent, provider, model, status, summary, created_at, updated_at)
VALUES ('run_1', ?, 'run_parent', 'chat', '', '', 'completed', 'done', ?, ?)`, sessionID, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `
INSERT INTO tool_events (session_id, tool_name, input, output, ok, error, created_at)
VALUES (?, 'web_search', '{"query":"redis sqlite"}', 'results...', 1, NULL, ?)`, sessionID, now); err != nil {
		t.Fatal(err)
	}

	detail, err := svc.SessionMessages(ctx, sessionID)
	if err != nil {
		t.Fatal(err)
	}
	if len(detail.ToolEvents) != 1 {
		t.Fatalf("tool_events len = %d", len(detail.ToolEvents))
	}
	if detail.ToolEvents[0].ToolName != "web_search" {
		t.Fatalf("tool name = %q", detail.ToolEvents[0].ToolName)
	}
	if detail.RunStatus != "completed" {
		t.Fatalf("run_status = %q", detail.RunStatus)
	}
}
