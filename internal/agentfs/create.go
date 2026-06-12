package agentfs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func WriteAgent(agentPath string, req CreateRequest) error {
	dir := filepath.Dir(agentPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, SkillsDirName), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, ToolsDirName), 0o755); err != nil {
		return err
	}

	spawnable := req.Spawnable
	if req.ID == "default" {
		spawnable = false
	}

	meta := fmt.Sprintf(`---
id: %s
name: %s
language: %s
tools: %s
providers: %s
model: %s
memory_scope: %s
max_steps: %d
spawnable: %t
auto_create: false
---
%s
`,
		req.ID,
		req.Name,
		req.Language,
		yamlStringList(req.Tools),
		yamlStringList(req.Providers),
		defaultString(req.Model, "inherit"),
		req.MemoryScope,
		req.MaxSteps,
		spawnable,
		strings.TrimSpace(req.Persona),
	)
	if err := os.WriteFile(agentPath, []byte(meta), 0o644); err != nil {
		return err
	}

	for i, skill := range req.Skills {
		name := strings.TrimSpace(skill.Name)
		if name == "" {
			name = fmt.Sprintf("skill-%d", i+1)
		}
		path := filepath.Join(dir, SkillsDirName, sanitizeFileName(name)+".md")
		content := fmt.Sprintf(`---
name: %s
triggers: %s
---
%s
`, name, yamlStringList(skill.Triggers), strings.TrimSpace(skill.Instructions))
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return err
		}
	}

	for _, guide := range req.ToolGuides {
		tool := strings.TrimSpace(guide.Tool)
		if tool == "" {
			continue
		}
		path := filepath.Join(dir, ToolsDirName, sanitizeFileName(tool)+".md")
		content := fmt.Sprintf(`---
type: guide
tool: %s
triggers: %s
---
%s
`, tool, yamlStringList(guide.Triggers), strings.TrimSpace(guide.Instructions))
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func CreateAgent(root string, req CreateRequest) (string, error) {
	if err := ValidateAgentID(req.ID); err != nil {
		return "", err
	}
	dir := filepath.Join(root, req.ID)
	agentPath := filepath.Join(dir, AgentFileName)
	if _, err := os.Stat(agentPath); err == nil {
		return "", fmt.Errorf("agent already exists: %s", req.ID)
	}
	if err := WriteAgent(agentPath, req); err != nil {
		return "", err
	}
	return dir, nil
}

func yamlStringList(items []string) string {
	if len(items) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(items))
	for _, item := range items {
		parts = append(parts, fmt.Sprintf("%q", item))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func sanitizeFileName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, " ", "-")
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		case r == '.':
			b.WriteRune('-')
		}
	}
	out := b.String()
	if out == "" {
		return "tool"
	}
	return out
}
