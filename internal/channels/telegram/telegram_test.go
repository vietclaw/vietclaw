package telegram

import (
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestTelegramMentionsBotFromEntity(t *testing.T) {
	msg := &tgbotapi.Message{
		Text: "ê @vietclawgo_bot ping",
		Entities: []tgbotapi.MessageEntity{{
			Type:   "mention",
			Offset: 2,
			Length: 16,
		}},
	}
	if !telegramMentionsBot(msg, 42, "@vietclawgo_bot", msg.Text) {
		t.Fatal("expected mention entity to match bot username")
	}
}

func TestTelegramMentionsBotFromUTF16Entity(t *testing.T) {
	msg := &tgbotapi.Message{
		Text: "😅 @vietclawgo_bot ping",
		Entities: []tgbotapi.MessageEntity{{
			Type:   "mention",
			Offset: 3,
			Length: 16,
		}},
	}
	if !telegramMentionsBot(msg, 42, "@vietclawgo_bot", msg.Text) {
		t.Fatal("expected UTF-16 entity offset to match bot username")
	}
}

func TestTelegramMentionsBotFromCommand(t *testing.T) {
	msg := &tgbotapi.Message{
		Text: "/start@vietclawgo_bot",
		Entities: []tgbotapi.MessageEntity{{
			Type:   "bot_command",
			Offset: 0,
			Length: 20,
		}},
	}
	if !telegramMentionsBot(msg, 42, "@vietclawgo_bot", msg.Text) {
		t.Fatal("expected bot command suffix to match bot username")
	}
}

func TestTelegramHTMLFormatsMarkdownSubset(t *testing.T) {
	got := telegramHTML("**CPU**\n`amd64`\n```go\nfmt.Println(\"x\")\n```")
	if !containsAll(got, "<b>CPU</b>", "<code>amd64</code>", "<pre>fmt.Println(&#34;x&#34;)\n</pre>") {
		t.Fatalf("unexpected telegram html: %s", got)
	}
}

func containsAll(text string, values ...string) bool {
	for _, value := range values {
		if !strings.Contains(text, value) {
			return false
		}
	}
	return true
}
