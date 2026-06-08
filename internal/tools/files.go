package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FileRead struct {
	Policy Policy
}

func (t FileRead) Name() string { return "file.read" }

func (t FileRead) Run(_ context.Context, input string) (string, error) {
	path, err := t.Policy.FileAllowed(input)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	return string(data), err
}

type FileWrite struct {
	Policy Policy
}

func (t FileWrite) Name() string { return "file.write" }

func (t FileWrite) Run(_ context.Context, input string) (string, error) {
	parts := strings.SplitN(input, "\n", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("file.write input must be path newline content")
	}
	path, err := t.Policy.FileAllowed(parts[0])
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(parts[1]), 0o644); err != nil {
		return "", err
	}
	return "ok", nil
}
