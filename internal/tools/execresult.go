package tools

import (
	"fmt"
	"os/exec"
	"strings"
)

// CombinedOutputResult pairs process output with a clearer error when the command fails.
func CombinedOutputResult(out []byte, err error) (string, error) {
	text := string(out)
	if err == nil {
		return text, nil
	}
	wrapped := wrapExecError(err)
	if trimmed := strings.TrimSpace(text); trimmed != "" {
		return text, wrapped
	}
	return "", wrapped
}

func wrapExecError(err error) error {
	if ee, ok := err.(*exec.ExitError); ok {
		return fmt.Errorf("exit code %d", ee.ExitCode())
	}
	return err
}
