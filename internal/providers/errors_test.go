package providers

import (
	"strings"
	"testing"
)

func TestSanitizeError(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Whitespaces are trimmed",
			input:    "   some error message   ",
			expected: "some error message",
		},
		{
			name:     "Normal message",
			input:    "Rate limit exceeded",
			expected: "Rate limit exceeded",
		},
		{
			name:     "Long message truncation",
			input:    strings.Repeat("a", 400),
			expected: strings.Repeat("a", maxProviderErrorLength),
		},
		{
			name:     "OpenAI key redaction (sk-)",
			input:    "Error: invalid api key sk-1234567890abcdef1234567890abcdef",
			expected: "Error: invalid api key [REDACTED]",
		},
		{
			name:     "OpenAI key redaction (sk-proj-)",
			input:    "Using key sk-proj-abCDef1234_xyz-5678",
			expected: "Using key [REDACTED]",
		},
		{
			name:     "Anthropic key redaction",
			input:    "Auth failed for sk-ant-api03-abcdef1234567890-xyz",
			expected: "Auth failed for [REDACTED]",
		},
		{
			name:     "Gemini key redaction",
			input:    "Invalid token AIzaSyB_XYZ123abcDEF456ghiJKL789mnoPQRstu",
			expected: "Invalid token [REDACTED]",
		},
		{
			name:     "Multiple keys redaction",
			input:    "Found sk-1234567890abcdef1234567890abcdef and AIzaSyB_XYZ123abcDEF456ghiJKL789mnoPQRstu",
			expected: "Found [REDACTED] and [REDACTED]",
		},
		{
			name: "Redaction before truncation",
			// Ensure that if a key spans the truncation boundary, it's redacted first
			input:    strings.Repeat("a", 290) + "sk-1234567890abcdef1234567890abcdef" + strings.Repeat("b", 50),
			expected: (strings.Repeat("a", 290) + "[REDACTED]" + strings.Repeat("b", 50))[:maxProviderErrorLength],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeError(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDefaultString(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		fallback string
		expected string
	}{
		{
			name:     "Use value when present",
			value:    "actual_value",
			fallback: "fallback_value",
			expected: "actual_value",
		},
		{
			name:     "Use fallback when empty",
			value:    "",
			fallback: "fallback_value",
			expected: "fallback_value",
		},
		{
			name:     "Use fallback when only spaces",
			value:    "   ",
			fallback: "fallback_value",
			expected: "fallback_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := defaultString(tt.value, tt.fallback)
			if result != tt.expected {
				t.Errorf("defaultString() = %v, want %v", result, tt.expected)
			}
		})
	}
}
