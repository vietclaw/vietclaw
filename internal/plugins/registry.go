package plugins

import "vietclaw/internal/framework"

// BuiltinRegistry returns the built-in extension catalog for the agent framework.
func BuiltinRegistry() []framework.ExtensionInfo {
	return framework.BuiltinExtensions()
}
