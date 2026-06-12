package agent_test

import (
	"strings"
	"testing"

	"vietclaw/internal/agent"
)

func TestParentSessionID(t *testing.T) {
	parent := "sess_abc123"
	child := parent + ":spawn:researcher:sub_deadbeef:delegate:researcher"
	if got := agent.ParentSessionID(child); got != parent {
		t.Fatalf("ParentSessionID = %q, want %q", got, parent)
	}
	if got := agent.ParentSessionID(parent); got != parent {
		t.Fatalf("root id = %q", got)
	}
}

func TestSpawnAgentID(t *testing.T) {
	child := "sess_abc:spawn:code-reviewer:sub_1122:delegate:code-reviewer"
	if got := agent.SpawnAgentID(child); got != "code-reviewer" {
		t.Fatalf("SpawnAgentID = %q", got)
	}
	if got := agent.SpawnAgentID("sess_only"); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestIsSpawnChildSession(t *testing.T) {
	if !agent.IsSpawnChildSession("sess:spawn:a:sub_x:delegate:a") {
		t.Fatal("expected spawn child")
	}
	if agent.IsSpawnChildSession("sess_root") {
		t.Fatal("expected root")
	}
}

func TestChildSessionIDFormat(t *testing.T) {
	parent := "sess_parent"
	agentID := "researcher"
	child := parent + ":spawn:" + agentID + ":sub_aaa:delegate:" + agentID
	if !strings.HasPrefix(child, parent+":spawn:") {
		t.Fatalf("bad prefix: %s", child)
	}
	if agent.ParentSessionID(child) != parent {
		t.Fatalf("parent mismatch")
	}
	if agent.SpawnAgentID(child) != agentID {
		t.Fatalf("agent mismatch")
	}
}
