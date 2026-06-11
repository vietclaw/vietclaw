package tools_test

import (
	"errors"
	"testing"

	"vietclaw/internal/tools"
)

func TestCombinedOutputResultSuccess(t *testing.T) {
	out, err := tools.CombinedOutputResult([]byte("hello\n"), nil)
	if err != nil || out != "hello\n" {
		t.Fatalf("got %q, %v", out, err)
	}
}

func TestCombinedOutputResultWithOutput(t *testing.T) {
	out, err := tools.CombinedOutputResult([]byte("traceback\n"), errors.New("exit status 1"))
	if err == nil {
		t.Fatal("expected error")
	}
	if out != "traceback\n" {
		t.Fatalf("output = %q", out)
	}
	if err.Error() != "exit status 1" {
		t.Fatalf("error = %v", err)
	}
}

func TestCombinedOutputResultNoOutput(t *testing.T) {
	out, err := tools.CombinedOutputResult(nil, errors.New("exit status 2"))
	if out != "" {
		t.Fatalf("output = %q", out)
	}
	if err == nil || err.Error() != "exit status 2" {
		t.Fatalf("error = %v", err)
	}
}
