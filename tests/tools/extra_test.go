package tools_test

import (
	"context"
	"strings"
	"testing"

	"vietclaw/internal/config"
	"vietclaw/internal/tools"
)

func TestExtraToolsRegistered(t *testing.T) {
	cfg := config.Default(config.Paths{DataDir: t.TempDir()})
	registry := tools.NewRegistry(cfg)
	defs := registry.GetDefinitions()

	names := map[string]bool{}
	for _, def := range defs {
		names[def.Function.Name] = true
	}
	want := []string{
		"uuid_generate", "random_string", "regex_extract", "regex_replace", "text_stats",
		"markdown_to_text", "csv_preview", "csv_to_json", "json_validate", "json_query",
		"url_parse", "html_to_text", "dns_lookup", "http_request", "timestamp_parse",
		"timestamp_format", "file_stat", "file_head", "file_tail", "path_info",
	}
	for _, name := range want {
		if !names[name] {
			t.Fatalf("missing extra tool definition %s", name)
		}
	}
}

func TestExtraToolExecution(t *testing.T) {
	registry := tools.NewRegistry(config.Default(config.Paths{DataDir: t.TempDir()}))
	out, err := registry.Execute(context.Background(), "regex_extract", `{"text":"abc 123 def 456","pattern":"\\d+"}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "123") || !strings.Contains(out, "456") {
		t.Fatalf("unexpected regex output: %s", out)
	}

	out, err = registry.Execute(context.Background(), "json_query", `{"text":"{\"user\":{\"name\":\"VietClaw\"}}","path":"user.name"}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "VietClaw") {
		t.Fatalf("unexpected json query output: %s", out)
	}
}
