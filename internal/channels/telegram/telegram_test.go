package telegram

import (
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
