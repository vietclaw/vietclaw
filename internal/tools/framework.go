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

func FrameworkToolDefinitions(allowAutoCreate bool) []providers.ToolDefinition {
	defs := []providers.ToolDefinition{
		AgentDelegateDefinition(),
		AgentSpawnDefinition(),
		AgentSpawnBatchDefinition(),
	}
	if allowAutoCreate {
		defs = append(defs, AgentCreateDefinition())
	}
	return defs
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
			Description: "Spawn multiple child agents in parallel and collect their results. Each task must use an agent_id whose specialty matches the message — do not assign research tasks to code-review agents or review tasks to research agents.",
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
			Name: ToolAgentCreate,
			Description: "Create a new specialized agent at ~/.vietclaw/agents/<id>/ with AGENT.md, skills/*.md, and tools/*.md guides. " +
				"REQUIRED: persona is the full AGENT.md markdown body in English OR Vietnamese (>=400 chars, >=3 ## sections). " +
				"Set language to vi or en in frontmatter; skills and tool_guides should use the same language. " +
				"REQUIRED: skills[] with >=1 detailed skill (instructions >=120 chars). " +
				"REQUIRED: tool_guides[] for every tool in tools[] (instructions >=80 chars each). " +
				"Do NOT create stub agents with 1-4 line descriptions — validation will reject them.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id":           map[string]any{"type": "string", "description": "Agent id: lowercase [a-z0-9][a-z0-9_-]*"},
					"name":         map[string]any{"type": "string", "description": "Human-readable display name"},
					"language":     map[string]any{"type": "string", "description": "vi or en — persona/skills/guides may be written in either language"},
					"persona":      map[string]any{"type": "string", "description": "Full AGENT.md markdown body (English or Vietnamese). Multi-section detailed spec (>=400 chars, >=3 ## headings)."},
					"tools":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Built-in tool names this agent may call"},
					"providers":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					"model":        map[string]any{"type": "string", "description": "inherit, catalog id, or provider/model"},
					"memory_scope": map[string]any{"type": "string"},
					"max_steps":    map[string]any{"type": "integer"},
					"spawnable":    map[string]any{"type": "boolean"},
					"skills": map[string]any{
						"type":        "array",
						"description": "At least one skill playbook (skills/<name>.md)",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"name":         map[string]any{"type": "string"},
								"triggers":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
								"instructions": map[string]any{"type": "string", "description": "Detailed skill instructions (>=120 chars)"},
							},
							"required": []string{"name", "instructions"},
						},
					},
					"tool_guides": map[string]any{
						"type":        "array",
						"description": "One guide per tool in tools[] (tools/<tool>.md)",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"tool":         map[string]any{"type": "string"},
								"triggers":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
								"instructions": map[string]any{"type": "string", "description": "When/how to use this tool (>=80 chars)"},
							},
							"required": []string{"tool", "instructions"},
						},
					},
				},
				"required": []string{"id", "persona", "skills", "tool_guides"},
			},
		},
	}
}
