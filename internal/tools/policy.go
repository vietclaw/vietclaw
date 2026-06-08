package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"vietclaw/internal/config"
)

type Policy struct {
	cfg config.Config
}

func NewPolicy(cfg config.Config) Policy {
	return Policy{cfg: cfg}
}

func (p Policy) ShellAllowed() bool {
	return p.cfg.Tools.Shell.Enabled
}

func (p Policy) FileAllowed(path string) (string, error) {
	if !p.cfg.Tools.Files.Enabled {
		return "", fmt.Errorf("file tools disabled")
	}
	workspace := config.ExpandPath(p.cfg.Agent.Workspace)
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		return "", err
	}
	cleaned := filepath.Clean(path)
	if !filepath.IsAbs(cleaned) {
		cleaned = filepath.Join(workspace, cleaned)
	}
	abs, err := filepath.Abs(cleaned)
	if err != nil {
		return "", err
	}
	workspaceAbs, err := filepath.Abs(workspace)
	if err != nil {
		return "", err
	}
	if p.cfg.Tools.Files.WorkspaceOnly {
		rel, err := filepath.Rel(workspaceAbs, abs)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			return "", fmt.Errorf("path outside workspace")
		}
	}
	return abs, nil
}
