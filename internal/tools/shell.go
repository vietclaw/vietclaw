package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type ShellExec struct {
	Policy Policy
}

func (t ShellExec) Name() string { return "shell.exec" }

func (t ShellExec) Run(ctx context.Context, input string) (string, error) {
	if !t.Policy.ShellAllowed() {
		return "", fmt.Errorf("shell.exec disabled")
	}
	fields := strings.Fields(input)
	if len(fields) == 0 {
		return "", fmt.Errorf("empty command")
	}
	out, err := exec.CommandContext(ctx, fields[0], fields[1:]...).CombinedOutput()
	return string(out), err
}
