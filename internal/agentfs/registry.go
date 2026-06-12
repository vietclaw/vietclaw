package agentfs

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"vietclaw/internal/config"
	"vietclaw/internal/skills"
)

type Registry struct {
	mu      sync.RWMutex
	root    string
	agents  map[string]AgentDefinition
	ordered []string
	cfg     config.Config
}

func NewRegistry(root string, cfg config.Config) *Registry {
	return &Registry{
		root:   root,
		agents: make(map[string]AgentDefinition),
		cfg:    cfg,
	}
}

func DefaultRoot(dataDir string) string {
	return filepath.Join(dataDir, DefaultAgentsDir)
}

func (r *Registry) Root() string {
	return r.root
}

func (r *Registry) Reload() error {
	if err := os.MkdirAll(r.root, 0o755); err != nil {
		return fmt.Errorf("create agents dir: %w", err)
	}
	entries, err := os.ReadDir(r.root)
	if err != nil {
		return err
	}

	loaded := make(map[string]AgentDefinition)
	var ordered []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		agentPath := filepath.Join(r.root, entry.Name(), AgentFileName)
		if _, err := os.Stat(agentPath); err != nil {
			continue
		}
		def, err := parseAgentFile(agentPath)
		if err != nil {
			continue
		}
		loaded[def.ID] = def
		ordered = append(ordered, def.ID)
	}

	if _, ok := loaded[config.DefaultAgentID]; !ok {
		def, err := r.ensureDefaultAgent()
		if err != nil {
			return err
		}
		loaded[def.ID] = def
		ordered = append([]string{def.ID}, ordered...)
	}

	r.mu.Lock()
	r.agents = loaded
	r.ordered = ordered
	r.mu.Unlock()
	return nil
}

func (r *Registry) ensureDefaultAgent() (AgentDefinition, error) {
	dir := filepath.Join(r.root, config.DefaultAgentID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return AgentDefinition{}, err
	}
	path := filepath.Join(dir, AgentFileName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := WriteAgent(path, CreateRequest{
			ID:        config.DefaultAgentID,
			Name:      r.cfg.Agent.Name,
			Language:  r.cfg.Agent.Language,
			Persona:   "",
			Tools:     []string{},
			Providers: []string{},
			Model:     "inherit",
			Spawnable: false,
		}); err != nil {
			return AgentDefinition{}, err
		}
	}
	return parseAgentFile(path)
}

func (r *Registry) Get(id string) (AgentDefinition, bool) {
	if id == "" {
		id = config.DefaultAgentID
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	def, ok := r.agents[id]
	return def, ok
}

func (r *Registry) List() []AgentDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]AgentDefinition, 0, len(r.ordered))
	for _, id := range r.ordered {
		if def, ok := r.agents[id]; ok {
			out = append(out, def)
		}
	}
	return out
}

func (r *Registry) Profiles() []config.AgentProfileConfig {
	list := r.List()
	out := make([]config.AgentProfileConfig, 0, len(list))
	for _, def := range list {
		out = append(out, def.Profile())
	}
	return out
}

func (r *Registry) SpawnableProfiles() []config.AgentProfileConfig {
	list := r.List()
	out := make([]config.AgentProfileConfig, 0)
	for _, def := range list {
		if def.Spawnable && def.ID != config.DefaultAgentID {
			out = append(out, def.Profile())
		}
	}
	return out
}

func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.agents)
}

func (r *Registry) SkillsFor(agentID string) []skills.Skill {
	def, ok := r.Get(agentID)
	if !ok {
		return nil
	}
	return def.Skills
}

func (r *Registry) ToolGuidesFor(agentID string) []ToolGuide {
	def, ok := r.Get(agentID)
	if !ok {
		return nil
	}
	return def.ToolGuides
}

func (r *Registry) CustomToolsFor(agentID string) []CustomTool {
	def, ok := r.Get(agentID)
	if !ok {
		return nil
	}
	return def.CustomTools
}
