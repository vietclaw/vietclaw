package contextbuilder

import (
	"fmt"
	"strings"

	"vietclaw/internal/config"
	"vietclaw/internal/i18n"
)

func (b *Builder) agentContextBlock(lang string) string {
	if b.agents == nil || !b.cfg.Framework.Enabled {
		return ""
	}

	lines := []string{i18n.T(lang, i18n.SystemAgentsHeader)}
	if b.cfg.Framework.AllowAutoCreate {
		lines = append(lines, i18n.T(lang, i18n.SystemAgentsAutoCreateEnabled))
		lines = append(lines, i18n.T(lang, i18n.SystemAgentsAutoCreateGuide))
	} else {
		lines = append(lines, i18n.T(lang, i18n.SystemAgentsAutoCreateDisabled))
	}
	lines = append(lines, i18n.T(lang, i18n.SystemAgentsSpawnRules))

	spawnable := 0
	for _, def := range b.agents.List() {
		if def.ID == config.DefaultAgentID {
			continue
		}
		spawnableFlag := "no"
		if def.Spawnable {
			spawnableFlag = "yes"
			spawnable++
		}
		name := strings.TrimSpace(def.Name)
		if name == "" {
			name = def.ID
		}
		persona := strings.TrimSpace(def.Persona)
		if persona != "" {
			persona = " — " + trimTo(persona, 120)
		}
		lines = append(lines, fmt.Sprintf("- %s (%s) spawnable=%s%s", def.ID, name, spawnableFlag, persona))
	}
	if spawnable == 0 {
		lines = append(lines, i18n.T(lang, i18n.SystemAgentsNoSpawnable))
	}
	return strings.Join(lines, "\n")
}
