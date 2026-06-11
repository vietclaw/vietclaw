package config

import "encoding/json"

// JsonPathExists reports whether a nested JSON object path is present in raw config bytes.
// Used to distinguish omitted fields from explicit zero/false values during merge.
func JsonPathExists(raw []byte, parts ...string) bool {
	if len(parts) == 0 || len(raw) == 0 {
		return false
	}
	var cur any
	if err := json.Unmarshal(raw, &cur); err != nil {
		return false
	}
	for _, part := range parts {
		m, ok := cur.(map[string]any)
		if !ok {
			return false
		}
		v, ok := m[part]
		if !ok {
			return false
		}
		cur = v
	}
	return true
}

func MergeAgentOptional(cfg AgentConfig, def AgentConfig, raw []byte) AgentConfig {
	if !JsonPathExists(raw, "agent", "max_steps") && cfg.MaxSteps == 0 {
		cfg.MaxSteps = def.MaxSteps
	}
	if !JsonPathExists(raw, "agent", "reflexion") {
		cfg.Reflexion = def.Reflexion
	}
	if !JsonPathExists(raw, "agent", "memory_tools") {
		cfg.MemoryTools = def.MemoryTools
	}
	if !JsonPathExists(raw, "agent", "heartbeat") {
		cfg.Heartbeat = def.Heartbeat
	}
	if !JsonPathExists(raw, "agent", "experience") {
		cfg.Experience = def.Experience
	}
	return cfg
}

func MergeFrameworkOptional(cfg FrameworkConfig, def FrameworkConfig, raw []byte) FrameworkConfig {
	if !JsonPathExists(raw, "framework") {
		return def
	}
	return cfg
}
