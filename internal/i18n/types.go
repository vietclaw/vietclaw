package i18n

type Language string

const (
	LanguageVietnamese Language = "vi"
	LanguageEnglish    Language = "en"
)

type MessageID string

const (
	AgentActionBlocked     MessageID = "agent.action_blocked"
	AgentMessageRequired   MessageID = "agent.message_required"
	MemorySaved            MessageID = "memory.saved"
	MemoryFound            MessageID = "memory.found"
	MemoryNotFound         MessageID = "memory.not_found"
	ChannelEmptyPrompt     MessageID = "channel.empty_prompt"
	ChannelEmptyAgentReply MessageID = "channel.empty_agent_reply"
	ProviderMockDefault    MessageID = "provider.mock.default"
	ProviderMockMemory     MessageID = "provider.mock.memory"
	SystemPromptBase       MessageID = "system.prompt_base"
	SystemMemoryHeader     MessageID = "system.memory_header"
	SystemHistoryHeader    MessageID = "system.history_header"
	CLIUsage               MessageID = "cli.usage"
	CLIErrorPrefix         MessageID = "cli.error_prefix"
	CLIMemorySaved         MessageID = "cli.memory_saved"
)
