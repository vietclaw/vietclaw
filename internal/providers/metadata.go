package providers

import "vietclaw/internal/config"

func metadataLanguage(metadata map[string]any) string {
	value, ok := metadata["language"].(string)
	if !ok || value == "" {
		return config.DefaultAgentLanguage
	}
	return value
}
