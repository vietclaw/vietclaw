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

func TestSessionChildrenListsSpawnSessions(t *testing.T) {
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

	parentID := "sess_parent_1"
	childID := parentID + ":spawn:researcher:sub_abc:delegate:researcher"
	task := "Find VPS prices in Vietnam"
	now := time.Now().UTC().Format(time.RFC3339)

	if _, err := database.ExecContext(ctx, `
INSERT INTO sessions (id, channel, user_id, title, summary, created_at, updated_at)
VALUES (?, 'web', 'local', NULL, NULL, ?, ?)`, parentID, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `
INSERT INTO sessions (id, channel, user_id, title, summary, created_at, updated_at)
VALUES (?, 'web', 'local', NULL, NULL, ?, ?)`, childID, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `
INSERT INTO messages (session_id, role, content, created_at)
VALUES (?, 'user', ?, ?)`, childID, task, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `
INSERT INTO agent_runs (id, session_id, parent_run_id, intent, provider, model, status, summary, created_at, updated_at)
VALUES ('run_1', ?, '', 'chat', '', '', 'completed', 'done', ?, ?)`, childID, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `
INSERT INTO messages (session_id, role, content, created_at)
VALUES (?, 'assistant', 'Here are VPS options', ?)`, childID, now); err != nil {
		t.Fatal(err)
	}

	children, err := svc.SessionChildren(ctx, parentID)
	if err != nil {
		t.Fatal(err)
	}
	if len(children) != 1 {
		t.Fatalf("children len = %d", len(children))
	}
	child := children[0]
	if child.ID != childID {
		t.Fatalf("child id = %q", child.ID)
	}
	if child.AgentID != "researcher" {
		t.Fatalf("agent id = %q", child.AgentID)
	}
	if child.TaskPreview != task {
		t.Fatalf("task_preview = %q", child.TaskPreview)
	}
	if child.RunStatus != "completed" {
		t.Fatalf("run_status = %q", child.RunStatus)
	}
	if !child.HasReply {
		t.Fatal("expected has_reply true")
	}
}
