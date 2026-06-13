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
	AgentMaxStepsReached   MessageID = "agent.max_steps_reached"
	AgentToolExecuteError  MessageID = "agent.tool_execute_error"
	AgentToolFailed        MessageID = "agent.tool_failed"
	AgentToolOutputSection MessageID = "agent.tool_output_section"
	AgentToolNoOutput      MessageID = "agent.tool_no_output"
	AgentToolStatus        MessageID = "agent.tool_status"
	AgentToolResult        MessageID = "agent.tool_result"
	ToolFileRead           MessageID = "tool.file_read"
	ToolFileWrite          MessageID = "tool.file_write"
	ToolShellExec          MessageID = "tool.shell_exec"
	ToolPathParam          MessageID = "tool.param.path"
	ToolContentParam       MessageID = "tool.param.content"
	ToolCommandParam       MessageID = "tool.param.command"
	ToolWebSearch          MessageID = "tool.web_search"
	ToolWebFetch           MessageID = "tool.web_fetch"
	ToolQueryParam         MessageID = "tool.param.query"
	ToolURLParam           MessageID = "tool.param.url"
	SystemWebResearchGuide MessageID = "system.web_research_guide"
	SystemPromptBase       MessageID = "system.prompt_base"
	SystemMemoryHeader     MessageID = "system.memory_header"
	SystemHistoryHeader    MessageID = "system.history_header"
	SystemSummaryPrefix    MessageID = "system.summary_prefix"
	SystemSummarizePrompt  MessageID = "system.summarize_prompt"
	SystemSkillHeader            MessageID = "system.skill_header"
	SystemAgentsHeader           MessageID = "system.agents_header"
	SystemAgentsAutoCreateEnabled  MessageID = "system.agents_auto_create_enabled"
	SystemAgentsAutoCreateDisabled MessageID = "system.agents_auto_create_disabled"
	SystemAgentsAutoCreateGuide    MessageID = "system.agents_auto_create_guide"
	SystemAgentsNoSpawnable        MessageID = "system.agents_no_spawnable"
	SystemAgentsSpawnRules         MessageID = "system.agents_spawn_rules"
	CLIUsage               MessageID = "cli.usage"
	CLIErrorPrefix         MessageID = "cli.error_prefix"
	CLIMemorySaved         MessageID = "cli.memory_saved"
)
