package agentfs

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"vietclaw/internal/config"
	"vietclaw/internal/skills"
)

var validAgentID = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

type agentMeta struct {
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Language    string   `yaml:"language"`
	Tools       []string `yaml:"tools"`
	Providers   []string `yaml:"providers"`
	Model       string   `yaml:"model"`
	MemoryScope string   `yaml:"memory_scope"`
	MaxSteps    int      `yaml:"max_steps"`
	Spawnable   *bool    `yaml:"spawnable"`
	AutoCreate  *bool    `yaml:"auto_create"`
}

type toolMeta struct {
	Type        string         `yaml:"type"`
	Tool        string         `yaml:"tool"`
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Triggers    []string       `yaml:"triggers"`
	Handler     string         `yaml:"handler"`
	Parameters  map[string]any `yaml:"parameters"`
}

func splitFrontmatter(data []byte) (metaYAML string, body string, ok bool) {
	text := string(data)
	text = strings.TrimPrefix(text, "\ufeff")
	if !strings.HasPrefix(text, "---") {
		return "", strings.TrimSpace(text), false
	}
	rest := strings.TrimPrefix(text, "---")
	rest = strings.TrimLeft(rest, " \t\r\n")
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return "", strings.TrimSpace(text), false
	}
	metaYAML = strings.TrimSpace(rest[:end])
	body = strings.TrimSpace(rest[end+4:])
	return metaYAML, body, true
}

func parseAgentFile(path string) (AgentDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return AgentDefinition{}, err
	}
	dir := filepath.Dir(path)
	dirName := filepath.Base(dir)

	var meta agentMeta
	body := strings.TrimSpace(string(data))
	if metaYAML, personaBody, ok := splitFrontmatter(data); ok {
		if err := yaml.Unmarshal([]byte(metaYAML), &meta); err != nil {
			return AgentDefinition{}, fmt.Errorf("parse %s frontmatter: %w", path, err)
		}
		if strings.TrimSpace(personaBody) != "" {
			body = personaBody
		}
	}

	id := strings.TrimSpace(meta.ID)
	if id == "" {
		id = dirName
	}
	if !validAgentID.MatchString(id) {
		return AgentDefinition{}, fmt.Errorf("invalid agent id %q in %s", id, path)
	}

	name := strings.TrimSpace(meta.Name)
	if name == "" {
		name = id
	}
	lang := strings.TrimSpace(meta.Language)
	if lang == "" {
		lang = config.DefaultAgentLanguage
	}

	spawnable := true
	if meta.Spawnable != nil {
		spawnable = *meta.Spawnable
	}
	autoCreate := false
	if meta.AutoCreate != nil {
		autoCreate = *meta.AutoCreate
	}

	def := AgentDefinition{
		ID:          id,
		Name:        name,
		Language:    lang,
		Persona:     body,
		Tools:       meta.Tools,
		Providers:   meta.Providers,
		Model:       strings.TrimSpace(meta.Model),
		MemoryScope: strings.TrimSpace(meta.MemoryScope),
		MaxSteps:    meta.MaxSteps,
		Spawnable:   spawnable,
		AutoCreate:  autoCreate,
		Dir:         dir,
	}

	def.Skills, _ = loadSkills(filepath.Join(dir, SkillsDirName))
	def.ToolGuides, def.CustomTools, _ = loadTools(filepath.Join(dir, ToolsDirName))
	return def, nil
}

func loadSkills(dir string) ([]skills.Skill, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []skills.Skill
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".md") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		skill, err := parseSkillFile(path)
		if err != nil {
			continue
		}
		out = append(out, skill)
	}
	return out, nil
}

func parseSkillFile(path string) (skills.Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return skills.Skill{}, err
	}
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

	var meta struct {
		Name        string   `yaml:"name"`
		Description string   `yaml:"description"`
		Triggers    []string `yaml:"triggers"`
	}
	body := strings.TrimSpace(string(data))
	if metaYAML, content, ok := splitFrontmatter(data); ok {
		_ = yaml.Unmarshal([]byte(metaYAML), &meta)
		if strings.TrimSpace(content) != "" {
			body = content
		}
	}

	name := strings.TrimSpace(meta.Name)
	if name == "" {
		name = base
	}
	return skills.Skill{
		Name:         name,
		Description:  meta.Description,
		Triggers:     meta.Triggers,
		Instructions: body,
		Path:         path,
	}, nil
}

func loadTools(dir string) ([]ToolGuide, []CustomTool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	var guides []ToolGuide
	var custom []CustomTool
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".md") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		g, c, err := parseToolFile(path)
		if err != nil {
			continue
		}
		if g != nil {
			guides = append(guides, *g)
		}
		if c != nil {
			custom = append(custom, *c)
		}
	}
	return guides, custom, nil
}

func parseToolFile(path string) (*ToolGuide, *CustomTool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

	var meta toolMeta
	body := strings.TrimSpace(string(data))
	if metaYAML, content, ok := splitFrontmatter(data); ok {
		if err := yaml.Unmarshal([]byte(metaYAML), &meta); err != nil {
			return nil, nil, err
		}
		if strings.TrimSpace(content) != "" {
			body = content
		}
	}

	toolType := strings.ToLower(strings.TrimSpace(meta.Type))
	if toolType == "" {
		if strings.TrimSpace(meta.Tool) != "" {
			toolType = "guide"
		} else if strings.TrimSpace(meta.Name) != "" {
			toolType = "custom"
		}
	}

	switch toolType {
	case "guide":
		toolName := strings.TrimSpace(meta.Tool)
		if toolName == "" {
			toolName = base
		}
		return &ToolGuide{
			Tool:         toolName,
			Triggers:     meta.Triggers,
			Instructions: body,
			Path:         path,
		}, nil, nil
	case "custom":
		name := strings.TrimSpace(meta.Name)
		if name == "" {
			name = base
		}
		return nil, &CustomTool{
			Name:         name,
			Description:  meta.Description,
			Parameters:   meta.Parameters,
			Handler:      strings.TrimSpace(meta.Handler),
			Instructions: body,
			Path:         path,
		}, nil
	default:
		return nil, nil, fmt.Errorf("unknown tool type %q", toolType)
	}
}

func ValidateAgentID(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("agent id is required")
	}
	if id == config.DefaultAgentID {
		return fmt.Errorf("cannot overwrite system agent %q", id)
	}
	if !validAgentID.MatchString(id) {
		return fmt.Errorf("invalid agent id %q", id)
	}
	return nil
}
