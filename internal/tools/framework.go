package tools

import "vietclaw/internal/providers"

const ToolAgentDelegate = "agent_delegate"

func AgentDelegateDefinition() providers.ToolDefinition {
	return providers.ToolDefinition{
		Type: "function",
		Function: providers.FunctionDetail{
			Name: ToolAgentDelegate,
			Description: "Delegate a sub-task to another agent profile. The child agent runs with its own persona, tools, and memory scope, then returns a summary.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"agent_id": map[string]any{
						"type":        "string",
						"description": "Target agent profile id (e.g. researcher).",
					},
					"message": map[string]any{
						"type":        "string",
						"description": "Task message for the delegated agent.",
					},
				},
				"required": []string{"agent_id", "message"},
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
