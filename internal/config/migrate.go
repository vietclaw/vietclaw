package config

// ApplyLegacyMigrations upgrades values from older VietClaw defaults that cut tasks
// mid-flight or truncate long replies. Runs on every daemon start; saves when changed.
func ApplyLegacyMigrations(cfg Config) Config {
	switch cfg.Agent.MaxSteps {
	case 5, 12:
		cfg.Agent.MaxSteps = DefaultMaxAgentSteps
	}
	if cfg.Agent.MaxOutputTokens > 0 {
		cfg.Agent.MaxOutputTokens = DefaultMaxOutputTokens
	}
	for i := range cfg.Agents {
		switch cfg.Agents[i].MaxSteps {
		case 5, 12:
			cfg.Agents[i].MaxSteps = 0
		}
	}
	return cfg
}
