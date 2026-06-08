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
