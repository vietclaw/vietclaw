package channels_test

import (
	"context"
	"path/filepath"
	"testing"

	"vietclaw/internal/channels"
	"vietclaw/internal/config"
	"vietclaw/internal/db"
)

func TestHandleModelCommand(t *testing.T) {
	database, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := db.ApplySchema(database); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default(config.Paths{DataDir: t.TempDir()})
	cfg.Models.Catalog = []config.CatalogModelConfig{{
		ID: "fast", Provider: "mock", Model: "mock-small", Label: "Fast", Enabled: true,
	}}
	cfg.Models.DefaultCatalogID = "fast"

	list, err := channels.HandleModelCommand(context.Background(), database, cfg, channels.PlatformTelegram, "chat-1", "/models")
	if err != nil {
		t.Fatal(err)
	}
	if !list.Handled || list.Reply == "" {
		t.Fatalf("list result = %#v", list)
	}

	set, err := channels.HandleModelCommand(context.Background(), database, cfg, channels.PlatformTelegram, "chat-1", "/models fast")
	if err != nil {
		t.Fatal(err)
	}
	if !set.Handled {
		t.Fatal("expected handled set")
	}
	if got := channels.ReadCatalogPreference(database, channels.PlatformTelegram, "chat-1"); got != "fast" {
		t.Fatalf("preference = %q", got)
	}
}
