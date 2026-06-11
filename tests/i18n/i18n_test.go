package i18n_test

import (
	"strings"
	"testing"

	"vietclaw/internal/i18n"
)

func TestVietnameseAndEnglishMessages(t *testing.T) {
	vi := i18n.T("vi", i18n.MemorySaved, "server chính")
	en := i18n.T("en", i18n.MemorySaved, "main server")

	if !strings.Contains(vi, "t lưu") {
		t.Fatalf("unexpected vi text: %q", vi)
	}
	if !strings.Contains(en, "saved") {
		t.Fatalf("unexpected en text: %q", en)
	}
}

func TestUnknownLanguageFallsBackToVietnamese(t *testing.T) {
	got := i18n.T("jp", i18n.MemoryNotFound)
	if !strings.Contains(got, "memory") {
		t.Fatalf("unexpected fallback text: %q", got)
	}
}

func TestToolUILabel(t *testing.T) {
	vi := i18n.ToolUILabel("vi", "web_fetch")
	if vi != "Đã truy cập" {
		t.Fatalf("vi web_fetch = %q", vi)
	}
	en := i18n.ToolUILabel("en", "web_fetch")
	if en != "Fetched page" {
		t.Fatalf("en web_fetch = %q", en)
	}
	if got := i18n.ToolUILabel("vi", "unknown_tool_xyz"); got != "unknown_tool_xyz" {
		t.Fatalf("unknown tool fallback = %q", got)
	}
}
