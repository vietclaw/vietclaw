package channels_test

import (
	"strings"
	"testing"

	"vietclaw/internal/channels"
	"vietclaw/internal/config"
)

func TestPromptWithAttachments(t *testing.T) {
	got := channels.PromptWithAttachments("review this", []channels.Attachment{{
		Name:        "main.go",
		ContentType: "text/x-go",
		Size:        12,
		Text:        "package main",
	}})
	if !strings.Contains(got, "review this") || !strings.Contains(got, "Attachment: main.go") || !strings.Contains(got, "package main") {
		t.Fatalf("prompt missing attachment content: %q", got)
	}
}

func TestAttachmentAllowedTextAndCode(t *testing.T) {
	cfg := config.Default(config.Paths{DataDir: t.TempDir()}).Channels.Attachments
	if !channels.AttachmentAllowed("notes.txt", "text/plain", 32, cfg) {
		t.Fatal("txt attachment should be allowed")
	}
	if !channels.AttachmentAllowed("main.go", "", 32, cfg) {
		t.Fatal("go attachment should be allowed by extension")
	}
	if channels.AttachmentAllowed("image.png", "image/png", 32, cfg) {
		t.Fatal("png attachment should not be allowed")
	}
	if channels.AttachmentAllowed("huge.txt", "text/plain", cfg.MaxBytes+1, cfg) {
		t.Fatal("oversized attachment should not be allowed")
	}
}
