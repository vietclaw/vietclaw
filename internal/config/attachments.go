package config

func DefaultTextAttachmentExtensions() []string {
	return []string{
		".txt", ".md", ".markdown", ".rst", ".log",
		".json", ".jsonl", ".yaml", ".yml", ".toml", ".xml", ".csv", ".tsv",
		".go", ".py", ".js", ".jsx", ".ts", ".tsx", ".vue", ".svelte",
		".java", ".kt", ".kts", ".cs", ".cpp", ".c", ".h", ".hpp",
		".rs", ".php", ".rb", ".swift", ".scala", ".sh", ".bash", ".zsh",
		".ps1", ".bat", ".cmd", ".sql", ".html", ".css", ".scss", ".less",
		".dockerfile", ".gitignore", ".env.example",
	}
}
