package tools

import "vietclaw/internal/providers"

const (
	ToolAgentDelegate    = "agent_delegate"
	ToolAgentSpawn       = "agent_spawn"
	ToolAgentSpawnBatch  = "agent_spawn_batch"
	ToolAgentCreate      = "agent_create"
)

func IsFrameworkTool(name string) bool {
	switch name {
	case ToolAgentDelegate, ToolAgentSpawn, ToolAgentSpawnBatch, ToolAgentCreate:
		return true
	default:
		return false
	}
}

func FrameworkToolDefinitions() []providers.ToolDefinition {
	return []providers.ToolDefinition{
		AgentDelegateDefinition(),
		AgentSpawnDefinition(),
		AgentSpawnBatchDefinition(),
		AgentCreateDefinition(),
	}
}

func AgentDelegateDefinition() providers.ToolDefinition {
	return providers.ToolDefinition{
		Type: "function",
		Function: providers.FunctionDetail{
			Name:        ToolAgentDelegate,
			Description: "Delegate a sub-task to another agent (sync). Alias for agent_spawn.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"agent_id": map[string]any{"type": "string", "description": "Target agent id."},
					"message":  map[string]any{"type": "string", "description": "Task message."},
					"model":    map[string]any{"type": "string", "description": "inherit, catalog id, or provider/model."},
				},
				"required": []string{"agent_id", "message"},
			},
		},
	}
}

func AgentSpawnDefinition() providers.ToolDefinition {
	return providers.ToolDefinition{
		Type: "function",
		Function: providers.FunctionDetail{
			Name:        ToolAgentSpawn,
			Description: "Spawn a child agent to run a specialized sub-task and return the result.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"agent_id": map[string]any{"type": "string", "description": "Target agent id."},
					"message":  map[string]any{"type": "string", "description": "Task message."},
					"model":    map[string]any{"type": "string", "description": "inherit, catalog id, or provider/model."},
				},
				"required": []string{"agent_id", "message"},
			},
		},
	}
}

func AgentSpawnBatchDefinition() providers.ToolDefinition {
	return providers.ToolDefinition{
		Type: "function",
		Function: providers.FunctionDetail{
			Name:        ToolAgentSpawnBatch,
			Description: "Spawn multiple child agents in parallel and collect their results.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"tasks": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"agent_id": map[string]any{"type": "string"},
								"message":  map[string]any{"type": "string"},
								"model":    map[string]any{"type": "string"},
							},
							"required": []string{"agent_id", "message"},
						},
					},
				},
				"required": []string{"tasks"},
			},
		},
	}
}

func filterDefinitions(defs []providers.ToolDefinition, allowed []string) []providers.ToolDefinition {
	if len(allowed) == 0 {
		return defs
	}
	allow := map[string]bool{}
	for _, name := range allowed {
		allow[normalizeToolName(name)] = true
	}
	out := make([]providers.ToolDefinition, 0, len(allowed))
	for _, def := range defs {
		if allow[def.Function.Name] {
			out = append(out, def)
		}
	}
	return out
}

func AgentCreateDefinition() providers.ToolDefinition {
	return providers.ToolDefinition{
		Type: "function",
		Function: providers.FunctionDetail{
			Name:        ToolAgentCreate,
			Description: "Create a new specialized agent directory with AGENT.md, skills, and tool guides.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id":           map[string]any{"type": "string"},
					"name":         map[string]any{"type": "string"},
					"language":     map[string]any{"type": "string"},
					"persona":      map[string]any{"type": "string"},
					"tools":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					"providers":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					"model":        map[string]any{"type": "string"},
					"memory_scope": map[string]any{"type": "string"},
					"max_steps":    map[string]any{"type": "integer"},
					"spawnable":    map[string]any{"type": "boolean"},
					"skills": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"name":         map[string]any{"type": "string"},
								"triggers":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
								"instructions": map[string]any{"type": "string"},
							},
						},
					},
					"tool_guides": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"tool":         map[string]any{"type": "string"},
								"triggers":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
								"instructions": map[string]any{"type": "string"},
							},
						},
					},
				},
				"required": []string{"id", "persona"},
			},
		},
	}
}
