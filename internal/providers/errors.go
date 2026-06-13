package providers

import (
	"regexp"
	"strings"
)

const maxProviderErrorLength = 300

var (
	// Pre-compile regular expressions for performance as per guidelines
	openAIKeyPattern    = regexp.MustCompile(`(sk-[a-zA-Z0-9]{20,}|sk-proj-[a-zA-Z0-9_-]+)`)
	anthropicKeyPattern = regexp.MustCompile(`(sk-ant-[a-zA-Z0-9_-]+)`)
	// Gemini keys start with AIzaSy and are typically 39 chars long
	geminiKeyPattern = regexp.MustCompile(`(AIzaSy[a-zA-Z0-9_-]{33,})`)
)

func SanitizeError(value string) string {
	value = strings.TrimSpace(value)

	// Redact known API key formats
	value = openAIKeyPattern.ReplaceAllString(value, "[REDACTED]")
	value = anthropicKeyPattern.ReplaceAllString(value, "[REDACTED]")
	value = geminiKeyPattern.ReplaceAllString(value, "[REDACTED]")

	if len(value) > maxProviderErrorLength {
		return value[:maxProviderErrorLength]
	}
	return value
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
