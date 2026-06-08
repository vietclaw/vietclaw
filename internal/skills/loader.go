package skills

import (
	"os"
	"path/filepath"
	"strings"
)

const skillFileName = "SKILL.md"

func Load(dirs []string) ([]Skill, error) {
	var out []Skill
	for _, dir := range dirs {
		if strings.TrimSpace(dir) == "" {
			continue
		}
		err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() || entry.Name() != skillFileName {
				return nil
			}
			skill, err := parseFile(path)
			if err != nil {
				return nil
			}
			out = append(out, skill)
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			return out, err
		}
	}
	return out, nil
}

func Match(all []Skill, message string, limit int) []Skill {
	if limit <= 0 {
		limit = 3
	}
	text := strings.ToLower(message)
	var matched []Skill
	for _, skill := range all {
		needles := append([]string{skill.Name}, skill.Triggers...)
		for _, needle := range needles {
			needle = strings.ToLower(strings.TrimSpace(needle))
			if needle != "" && strings.Contains(text, needle) {
				matched = append(matched, skill)
				break
			}
		}
		if len(matched) >= limit {
			break
		}
	}
	return matched
}
