package context_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"vietclaw/internal/config"
	contextbuilder "vietclaw/internal/context"
	"vietclaw/internal/db"
	"vietclaw/internal/memory"
)

func TestHistorySkipsSystemToolLogs(t *testing.T) {
	database, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := db.ApplySchema(database); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default(config.Paths{DataDir: t.TempDir()})
	sessionID := "sess-history-filter"
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = database.Exec(`INSERT INTO sessions (id, channel, user_id, title, summary, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		sessionID, "web", "local", nil, nil, now, now)
	if err != nil {
		t.Fatal(err)
	}
	_, err = database.Exec(`INSERT INTO messages (session_id, role, content, created_at) VALUES (?, ?, ?, ?)`,
		sessionID, "user", "hello", now)
	if err != nil {
		t.Fatal(err)
	}
	_, err = database.Exec(`INSERT INTO messages (session_id, role, content, created_at) VALUES (?, ?, ?, ?)`,
		sessionID, "system", "[Tool Execution: file_read]\nInput: {}\nOutput: ok", now)
	if err != nil {
		t.Fatal(err)
	}
	_, err = database.Exec(`INSERT INTO messages (session_id, role, content, created_at) VALUES (?, ?, ?, ?)`,
		sessionID, "assistant", "done", now)
	if err != nil {
		t.Fatal(err)
	}

	builder := contextbuilder.New(cfg, database, memory.NewStore(database))
	messages, err := builder.Messages(context.Background(), sessionID, "user:local", config.DefaultAgentID, "hello", nil)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(messages[0].Content, "[Tool Execution:") {
		t.Fatalf("system tool logs leaked into context: %s", messages[0].Content)
	}
	if !strings.Contains(messages[0].Content, "hello") && !strings.Contains(messages[0].Content, "done") {
		t.Fatalf("expected user/assistant history in context: %s", messages[0].Content)
	}
}
