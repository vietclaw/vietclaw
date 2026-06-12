package agentfs

import (
	"vietclaw/internal/config"
	"vietclaw/internal/skills"
)

const (
	AgentFileName   = "AGENT.md"
	SkillsDirName   = "skills"
	ToolsDirName    = "tools"
	MigrationKey    = "agents_migrated_v1"
	DefaultAgentsDir = "agents"
)

type AgentDefinition struct {
	ID          string
	Name        string
	Language    string
	Persona     string
	Tools       []string
	Providers   []string
	Model       string
	MemoryScope string
	MaxSteps    int
	Spawnable   bool
	AutoCreate  bool
	Dir         string
	Skills      []skills.Skill
	ToolGuides  []ToolGuide
	CustomTools []CustomTool
}

func (a AgentDefinition) Profile() config.AgentProfileConfig {
	return config.AgentProfileConfig{
		ID:          a.ID,
		Name:        a.Name,
		Language:    a.Language,
		Persona:     a.Persona,
		Tools:       a.Tools,
		Providers:   a.Providers,
		MemoryScope: a.MemoryScope,
		MaxSteps:    a.MaxSteps,
	}
}

type ToolGuide struct {
	Tool         string
	Triggers     []string
	Instructions string
	Path         string
}

type CustomTool struct {
	Name        string
	Description string
	Parameters  map[string]any
	Handler     string
	Instructions string
	Path        string
}

type CreateRequest struct {
	ID          string
	Name        string
	Language    string
	Persona     string
	Tools       []string
	Providers   []string
	Model       string
	MemoryScope string
	MaxSteps    int
	Spawnable   bool
	Skills      []SkillInput
	ToolGuides  []ToolGuideInput
}

type SkillInput struct {
	Name         string
	Triggers     []string
	Instructions string
}

type ToolGuideInput struct {
	Tool         string
	Triggers     []string
	Instructions string
}
