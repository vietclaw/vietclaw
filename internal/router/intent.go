package router

import "strings"

type Intent string

const (
	IntentMemoryAdd   Intent = "memory_add"
	IntentMemoryQuery Intent = "memory_query"
	IntentChat        Intent = "chat"
	IntentAction      Intent = "action"
	IntentUnknown     Intent = "unknown"
)

func Classify(message string) Intent {
	text := strings.ToLower(strings.TrimSpace(message))
	switch {
	case text == "":
		return IntentUnknown
	case strings.Contains(text, "nhớ là") || strings.Contains(text, "từ nay") || strings.Contains(text, "lưu lại"):
		return IntentMemoryAdd
	case strings.Contains(text, "mày nhớ gì") || strings.Contains(text, "nhớ gì") || strings.Contains(text, "server chính") || strings.Contains(text, "đã lưu"):
		return IntentMemoryQuery
	case strings.HasPrefix(text, "chạy ") || strings.Contains(text, "đọc file") || strings.Contains(text, "ghi file"):
		return IntentAction
	default:
		return IntentChat
	}
}
