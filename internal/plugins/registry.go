package plugins

type Builtins struct {
	Providers []string `json:"providers"`
	Tools     []string `json:"tools"`
	Channels  []string `json:"channels"`
}

func BuiltinRegistry() Builtins {
	return Builtins{
		Providers: []string{"mock", "openai", "openai-compatible", "anthropic", "gemini", "http", "opencode-cli"},
		Tools:     []string{"file_read", "file_write", "shell_exec", "mcp"},
		Channels:  []string{"discord", "telegram"},
	}
}
