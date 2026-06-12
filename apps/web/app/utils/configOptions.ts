/** Known config enum values — must match backend (internal/config, internal/router). */

export const AGENT_EXPERIENCES = ['prompt', 'pro'] as const
export const AGENT_STYLES = ['natural_short'] as const
export const ROUTER_INTENT_MODES = ['rule', 'hybrid', 'llm'] as const
export const ROUTER_AGENT_ROUTINGS = ['rule', 'hybrid', 'llm'] as const
export const RUNTIME_MODES = ['eco', 'normal'] as const
export const SHELL_SANDBOXES = ['none', 'docker'] as const
export const WORKSPACE_MODES = ['ro', 'rw'] as const
export const MEMORY_SCOPES = ['global', 'user', 'session'] as const
export const MEMORY_KINDS = ['note', 'fact', 'rule', 'preference'] as const
export const MEMORY_CONFIDENCE = ['high', 'medium', 'low'] as const
export const UI_LANGUAGES = ['vi', 'en'] as const
export const TELEGRAM_COMMAND_MODES = ['slash', 'prefix'] as const

export type OptionGroup =
  | 'experience'
  | 'style'
  | 'intent_mode'
  | 'agent_routing'
  | 'runtime_mode'
  | 'sandbox'
  | 'workspace_mode'
  | 'memory_scope'
  | 'memory_kind'
  | 'memory_confidence'
  | 'language'
  | 'telegram_command_mode'

export const OPTION_GROUPS: Record<OptionGroup, readonly string[]> = {
  experience: AGENT_EXPERIENCES,
  style: AGENT_STYLES,
  intent_mode: ROUTER_INTENT_MODES,
  agent_routing: ROUTER_AGENT_ROUTINGS,
  runtime_mode: RUNTIME_MODES,
  sandbox: SHELL_SANDBOXES,
  workspace_mode: WORKSPACE_MODES,
  memory_scope: MEMORY_SCOPES,
  memory_kind: MEMORY_KINDS,
  memory_confidence: MEMORY_CONFIDENCE,
  language: UI_LANGUAGES,
  telegram_command_mode: TELEGRAM_COMMAND_MODES,
}
