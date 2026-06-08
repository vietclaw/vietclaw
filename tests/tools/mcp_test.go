package tools_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"vietclaw/internal/config"
	"vietclaw/internal/tools"
)

func TestMCPToolDiscoveryAndExecution(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method string `json:"method"`
			Params struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments"`
			} `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		switch req.Method {
		case "tools/list":
			_, _ = fmt.Fprint(w, `{"jsonrpc":"2.0","id":1,"result":{"tools":[{"name":"echo","description":"Echo text","inputSchema":{"type":"object","properties":{"text":{"type":"string"}}}}]}}`)
		case "tools/call":
			_, _ = fmt.Fprintf(w, `{"jsonrpc":"2.0","id":1,"result":{"content":[{"type":"text","text":"echo:%s"}]}}`, req.Params.Arguments["text"])
		default:
			t.Fatalf("unexpected method: %s", req.Method)
		}
	}))
	defer server.Close()

	cfg := config.Default(config.Paths{DataDir: t.TempDir()})
	cfg.Tools.MCP = []config.MCPServerConfig{{ID: "test", Enabled: true, URL: server.URL}}

	registry := tools.NewRegistry(cfg)
	defs := registry.GetDefinitions()
	found := false
	for _, def := range defs {
		if def.Function.Name == "mcp_test_echo" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected mcp tool definition, got %#v", defs)
	}

	out, err := registry.Execute(context.Background(), "mcp_test_echo", `{"text":"hello"}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "echo:hello") {
		t.Fatalf("unexpected mcp output: %q", out)
	}
}
