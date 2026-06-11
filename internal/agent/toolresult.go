package agent

import (
	"strings"

	"vietclaw/internal/i18n"
)

func formatToolFailureMessage(lang string, output string, execErr error) string {
	header := i18n.T(lang, i18n.AgentToolFailed, execErr.Error())
	trimmed := strings.TrimSpace(output)
	if trimmed != "" {
		return header + "\n" + i18n.T(lang, i18n.AgentToolOutputSection) + "\n" + trimmed
	}
	return header + "\n" + i18n.T(lang, i18n.AgentToolNoOutput)
}
