package tools_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"vietclaw/internal/config"
	"vietclaw/internal/tools"
)

func TestEvaluateMath(t *testing.T) {
	testCases := []struct {
		expr    string
		want    float64
		wantErr bool
	}{
		{"2 + 3", 5, false},
		{"10 - 4", 6, false},
		{"3 * 4", 12, false},
		{"20 / 5", 4, false},
		{"2 + 3 * 4", 14, false},
		{"(2 + 3) * 4", 20, false},
		{"-3.5 * 2", -7, false},
		{"10 / 0", 0, true},
		{"2 + (3 * (4 - 1))", 11, false},
		{"invalid", 0, true},
	}

	for _, tc := range testCases {
		got, err := tools.EvaluateMath(tc.expr)
		if tc.wantErr {
			if err == nil {
				t.Errorf("expected error for %s, got nil", tc.expr)
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error for %s: %v", tc.expr, err)
			}
			if got != tc.want {
				t.Errorf("EvaluateMath(%s) = %g, want %g", tc.expr, got, tc.want)
			}
		}
	}
}

func TestStringTransform(t *testing.T) {
	st := tools.StringTransform{}

	// Base64 encode
	res, err := st.Run(context.Background(), `{"text": "hello", "op": "base64_encode"}`)
	if err != nil {
		t.Fatal(err)
	}
	if res != "aGVsbG8=" {
		t.Errorf("expected aGVsbG8=, got %s", res)
	}

	// Base64 decode
	res, err = st.Run(context.Background(), `{"text": "aGVsbG8=", "op": "base64_decode"}`)
	if err != nil {
		t.Fatal(err)
	}
	if res != "hello" {
		t.Errorf("expected hello, got %s", res)
	}

	// URL encode
	res, err = st.Run(context.Background(), `{"text": "hello world!", "op": "url_encode"}`)
	if err != nil {
		t.Fatal(err)
	}
	if res != "hello+world%21" {
		t.Errorf("expected hello+world%%21, got %s", res)
	}

	// URL decode
	res, err = st.Run(context.Background(), `{"text": "hello+world%21", "op": "url_decode"}`)
	if err != nil {
		t.Fatal(err)
	}
	if res != "hello world!" {
		t.Errorf("expected hello world!, got %s", res)
	}

	// Upper / Lower
	res, err = st.Run(context.Background(), `{"text": "Hello", "op": "upper"}`)
	if err != nil {
		t.Fatal(err)
	}
	if res != "HELLO" {
		t.Errorf("expected HELLO, got %s", res)
	}

	res, err = st.Run(context.Background(), `{"text": "Hello", "op": "lower"}`)
	if err != nil {
		t.Fatal(err)
	}
	if res != "hello" {
		t.Errorf("expected hello, got %s", res)
	}
}

func TestJSONFormat(t *testing.T) {
	jf := tools.JSONFormat{}

	res, err := jf.Run(context.Background(), `{"text": "{\"b\":2,\"a\":1}", "minify": false}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res, "\n") || !strings.Contains(res, "  \"a\": 1") {
		t.Errorf("expected pretty printed json, got:\n%s", res)
	}

	res, err = jf.Run(context.Background(), `{"text": "{\n  \"b\": 2,\n  \"a\": 1\n}", "minify": true}`)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(res, "\n") {
		t.Errorf("expected compacted json, got:\n%s", res)
	}

	// Array
	resArr, err := jf.Run(context.Background(), `{"text": "[{\"a\":1},{\"b\":2}]", "minify": false}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resArr, "\n") {
		t.Errorf("expected pretty printed array, got:\n%s", resArr)
	}
}

func TestEnvGetSecurity(t *testing.T) {
	eg := tools.EnvGet{}
	os.Setenv("MY_APP_API_KEY", "sensitive_secret_123")
	defer os.Unsetenv("MY_APP_API_KEY")

	_, err := eg.Run(context.Background(), `{"key": "MY_APP_API_KEY"}`)
	if err == nil {
		t.Error("expected error/blocked access for sensitive env var, got nil")
	}

	os.Setenv("TEST_NORMAL_VAR", "value_here")
	defer os.Unsetenv("TEST_NORMAL_VAR")

	res, err := eg.Run(context.Background(), `{"key": "TEST_NORMAL_VAR"}`)
	if err != nil {
		t.Fatal(err)
	}
	if res != "value_here" {
		t.Errorf("expected value_here, got %s", res)
	}
}

func TestFileAndHashUtilities(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.Default(config.Paths{DataDir: tempDir})
	cfg.Tools.Files.Enabled = true
	cfg.Tools.Files.WorkspaceOnly = true
	cfg.Agent.Workspace = tempDir

	p := tools.NewPolicy(cfg)

	// Create test file
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "hello world from vietclaw tools"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Test HashCalc
	hc := tools.HashCalc{Policy: p}
	args := `{"path": "test.txt", "algo": "md5"}`
	res, err := hc.Run(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	expectedMD5 := "37396fa2b8da71d52049c5b873041042"
	if res != expectedMD5 {
		t.Errorf("expected MD5 %s, got %s", expectedMD5, res)
	}

	// Test DirList
	dl := tools.DirList{Policy: p}
	dlRes, err := dl.Run(context.Background(), `{"path": "."}`)
	if err != nil {
		t.Fatal(err)
	}
	var entries []tools.DirEntryInfo
	if err := json.Unmarshal([]byte(dlRes), &entries); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, entry := range entries {
		if entry.Name == "test.txt" {
			found = true
			if entry.IsDir {
				t.Error("test.txt should not be marked as directory")
			}
			if entry.Size != int64(len(testContent)) {
				t.Errorf("expected size %d, got %d", len(testContent), entry.Size)
			}
		}
	}
	if !found {
		t.Error("test.txt not found in directory listing")
	}

	// Test FileGrep
	fg := tools.FileGrep{Policy: p}
	fgRes, err := fg.Run(context.Background(), `{"path": "test.txt", "pattern": "vietclaw"}`)
	if err != nil {
		t.Fatal(err)
	}
	var grepMatches []tools.GrepMatch
	if err := json.Unmarshal([]byte(fgRes), &grepMatches); err != nil {
		t.Fatal(err)
	}
	if len(grepMatches) != 1 || grepMatches[0].Content != testContent {
		t.Errorf("expected 1 grep match with content, got: %v", grepMatches)
	}

	// Test FileFind
	ff := tools.FileFind{Policy: p}
	ffRes, err := ff.Run(context.Background(), `{"path": ".", "pattern": "*.txt"}`)
	if err != nil {
		t.Fatal(err)
	}
	var findMatches []string
	if err := json.Unmarshal([]byte(ffRes), &findMatches); err != nil {
		t.Fatal(err)
	}
	if len(findMatches) != 1 || filepath.Base(findMatches[0]) != "test.txt" {
		t.Errorf("expected test.txt in find matches, got: %v", findMatches)
	}
}

func TestWorkspaceAliasPaths(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.Default(config.Paths{DataDir: tempDir})
	cfg.Tools.Files.Enabled = true
	cfg.Tools.Files.WorkspaceOnly = true
	cfg.Agent.Workspace = tempDir
	p := tools.NewPolicy(cfg)

	for _, input := range []string{".", "workspace", "/workspace", "workspace/config.json", "/workspace/config.json"} {
		got, err := p.FileAllowed(input)
		if err != nil {
			t.Fatalf("FileAllowed(%q): %v", input, err)
		}
		if strings.Contains(got, filepath.Join("workspace", "workspace")) {
			t.Fatalf("workspace alias duplicated for %q: %s", input, got)
		}
	}
}
