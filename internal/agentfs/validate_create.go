package agentfs

import (
	"fmt"
	"strings"
)

const (
	MinPersonaRunes          = 400
	MinPersonaSections       = 3
	MinSkillInstructions     = 120
	MinToolGuideInstructions = 80
)

var frameworkOnlyTools = map[string]bool{
	"agent_create":      true,
	"agent_spawn":       true,
	"agent_spawn_batch": true,
	"agent_delegate":    true,
}

// ValidateCreateRequest ensures agent_create payloads include a detailed AGENT.md body.
func ValidateCreateRequest(req CreateRequest) error {
	if err := ValidateAgentID(req.ID); err != nil {
		return err
	}

	persona := strings.TrimSpace(req.Persona)
	if persona == "" {
		return fmt.Errorf("persona is required: write the full AGENT.md markdown body (not a one-line summary)")
	}
	if runeLen(persona) < MinPersonaRunes {
		return fmt.Errorf("persona too short (%d chars, need >= %d): expand AGENT.md in English or Vietnamese with role, workflow, output format, tool rules, limits, and examples", runeLen(persona), MinPersonaRunes)
	}
	if countMarkdownSections(persona) < MinPersonaSections {
		return fmt.Errorf("persona needs >= %d markdown sections (## headings), e.g. Role, Workflow, Output format, Tool rules, Limits, Examples", MinPersonaSections)
	}

	if len(req.Skills) == 0 {
		return fmt.Errorf("skills[] is required: add at least one skill with detailed instructions (playbook for common tasks)")
	}
	for i, skill := range req.Skills {
		if strings.TrimSpace(skill.Name) == "" {
			return fmt.Errorf("skills[%d].name is required", i)
		}
		if runeLen(strings.TrimSpace(skill.Instructions)) < MinSkillInstructions {
			return fmt.Errorf("skills[%d].instructions too short (need >= %d chars): describe steps, checks, and output for %q", i, MinSkillInstructions, skill.Name)
		}
	}

	for _, tool := range req.Tools {
		normalized := strings.TrimSpace(strings.ToLower(tool))
		if frameworkOnlyTools[normalized] {
			continue
		}
		guide, ok := findToolGuide(req.ToolGuides, normalized)
		if !ok {
			return fmt.Errorf("tool_guides entry required for tool %q: explain when and how to use it", tool)
		}
		if runeLen(strings.TrimSpace(guide.Instructions)) < MinToolGuideInstructions {
			return fmt.Errorf("tool_guides for %q too short (need >= %d chars)", tool, MinToolGuideInstructions)
		}
	}

	return nil
}

func runeLen(s string) int {
	return len([]rune(s))
}

func countMarkdownSections(text string) int {
	count := 0
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "##") {
			count++
		}
	}
	return count
}

func findToolGuide(guides []ToolGuideInput, tool string) (ToolGuideInput, bool) {
	for _, guide := range guides {
		if strings.TrimSpace(strings.ToLower(guide.Tool)) == tool {
			return guide, true
		}
	}
	return ToolGuideInput{}, false
}
