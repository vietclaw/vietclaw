package tools_test

import (
	"context"
	"os"
	"testing"

	"vietclaw/internal/tools"
)

func TestWebSearchLiveVietnameseQueries(t *testing.T) {
	if os.Getenv("VIETCLAW_LIVE_SEARCH") != "1" {
		t.Skip("set VIETCLAW_LIVE_SEARCH=1 to run live DuckDuckGo search test")
	}
	ws := tools.WebSearch{}
	queries := []string{
		"VPS gia re Viet Nam",
		"gia VPS Viet Nam 2024",
		"VPS Vietnam pricing",
	}
	for _, q := range queries {
		out, err := ws.Run(context.Background(), `{"query":"`+q+`"}`)
		if err != nil {
			t.Fatalf("query %q failed: %v", q, err)
		}
		t.Logf("%s => len=%d empty=%v", q, len(out), out == "[]")
		if out == "[]" {
			t.Fatalf("query %q returned no results", q)
		}
	}
}
