package agent

import (
	"errors"
	"strings"
	"testing"
)

func TestFormatToolFailureMessageWithOutput(t *testing.T) {
	msg := formatToolFailureMessage("vi", "line1\nline2", errors.New("exit code 1"))
	if !strings.Contains(msg, "Lệnh thất bại: exit code 1") {
		t.Fatalf("missing header: %q", msg)
	}
	if !strings.Contains(msg, "--- stdout/stderr ---") {
		t.Fatalf("missing section: %q", msg)
	}
	if !strings.Contains(msg, "line1") {
		t.Fatalf("missing output: %q", msg)
	}
}

func TestFormatToolFailureMessageNoOutput(t *testing.T) {
	msg := formatToolFailureMessage("en", "", errors.New("exit code 1"))
	if !strings.Contains(msg, "Command failed: exit code 1") {
		t.Fatalf("missing header: %q", msg)
	}
	if !strings.Contains(msg, "(no stdout/stderr captured)") {
		t.Fatalf("missing no-output hint: %q", msg)
	}
}
