package i18n

import (
	"fmt"
	"strings"
)

var catalog = map[Language]map[MessageID]string{
	LanguageVietnamese: {
		AgentActionBlocked:     "tool action cần policy rõ hơn. shell.exec đang tắt mặc định nếu chưa bật trong config.",
		AgentMessageRequired:   "message là bắt buộc",
		MemorySaved:            "ok, t lưu: %s",
		MemoryFound:            "t nhớ: %s",
		MemoryNotFound:         "t chưa thấy memory nào khớp.",
		ChannelEmptyPrompt:     "gọi t rồi muốn t làm gì?",
		ChannelEmptyAgentReply: "t chưa có gì để trả lời.",
		ProviderMockDefault:    "t là VietClaw, agent runtime nhẹ để điều phối model, memory và tools.",
		ProviderMockMemory:     "mock đây: t có thể lưu và tìm memory qua SQLite.",
		SystemPromptBase:       "Bạn là %s, agent điều phối nhẹ. Trả lời tiếng Việt ngắn, tự nhiên.",
		SystemMemoryHeader:     "Memory liên quan:",
		SystemHistoryHeader:    "Lịch sử gần đây:",
		CLIUsage:               "usage: vietclaw <version|init|daemon|status|doctor|chat|memory|channels|discord|telegram>",
		CLIErrorPrefix:         "lỗi:",
		CLIMemorySaved:         "ok, t lưu: %s",
	},
	LanguageEnglish: {
		AgentActionBlocked:     "tool action needs a clearer policy. shell.exec is disabled by default unless enabled in config.",
		AgentMessageRequired:   "message is required",
		MemorySaved:            "ok, saved: %s",
		MemoryFound:            "I remember: %s",
		MemoryNotFound:         "I did not find any matching memory.",
		ChannelEmptyPrompt:     "you called me, but what should I do?",
		ChannelEmptyAgentReply: "I do not have anything to reply with yet.",
		ProviderMockDefault:    "I am VietClaw, a lightweight agent runtime for routing models, memory, and tools.",
		ProviderMockMemory:     "mock here: I can save and search memory through SQLite.",
		SystemPromptBase:       "You are %s, a lightweight orchestration agent. Reply in concise, natural English.",
		SystemMemoryHeader:     "Relevant memory:",
		SystemHistoryHeader:    "Recent history:",
		CLIUsage:               "usage: vietclaw <version|init|daemon|status|doctor|chat|memory|channels|discord|telegram>",
		CLIErrorPrefix:         "error:",
		CLIMemorySaved:         "ok, saved: %s",
	},
}

func Normalize(language string) Language {
	switch Language(strings.ToLower(strings.TrimSpace(language))) {
	case LanguageEnglish:
		return LanguageEnglish
	default:
		return LanguageVietnamese
	}
}

func T(language string, id MessageID, args ...any) string {
	lang := Normalize(language)
	template := catalog[lang][id]
	if template == "" {
		template = catalog[LanguageVietnamese][id]
	}
	if template == "" {
		return string(id)
	}
	if len(args) == 0 {
		return template
	}
	return fmt.Sprintf(template, args...)
}
