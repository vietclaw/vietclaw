package skills

import (
	"os"
	"path/filepath"
	"strings"
)

func parseFile(path string) (Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Skill{}, err
	}
	text := string(data)
	lines := strings.Split(text, "\n")
	skill := Skill{
		Name:         filepath.Base(filepath.Dir(path)),
		Instructions: strings.TrimSpace(text),
		Path:         path,
	}
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		lower := strings.ToLower(line)
		switch {
		case strings.HasPrefix(line, "#") && skill.Name == filepath.Base(filepath.Dir(path)):
			skill.Name = strings.TrimSpace(strings.TrimLeft(line, "#"))
		case strings.HasPrefix(lower, "description:"):
			skill.Description = strings.TrimSpace(line[len("description:"):])
		case strings.HasPrefix(lower, "triggers:"):
			skill.Triggers = splitCSV(line[len("triggers:"):])
		case skill.Description == "" && line != "" && !strings.HasPrefix(line, "#"):
			skill.Description = line
		}
	}
	return skill, nil
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
