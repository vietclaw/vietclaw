package telegram

import (
	"context"
	"fmt"
	"html"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"vietclaw/internal/channels"
	"vietclaw/internal/config"
	"vietclaw/internal/i18n"
)

type Adapter struct {
	cfg        config.TelegramConfig
	attachment config.AttachmentConfig
	handler    *channels.Handler
}

func New(cfg config.TelegramConfig, attachment config.AttachmentConfig, handler *channels.Handler) *Adapter {
	return &Adapter{cfg: cfg, attachment: attachment, handler: handler}
}

func (a *Adapter) Name() string {
	return channels.PlatformTelegram
}

func (a *Adapter) Start(ctx context.Context) error {
	token := strings.TrimSpace(os.Getenv(a.cfg.TokenEnv))
	if token == "" {
		return fmt.Errorf("telegram token env missing: %s", a.cfg.TokenEnv)
	}
	_ = tgbotapi.SetLogger(redactingTelegramLogger{token: token, handler: a.handler})
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return fmt.Errorf("create telegram bot: %s", redactToken(err.Error(), token))
	}
	updateConfig := tgbotapi.NewUpdate(0)
	if a.cfg.PollTimeoutSeconds > 0 {
		updateConfig.Timeout = a.cfg.PollTimeoutSeconds
	}
	updates := bot.GetUpdatesChan(updateConfig)
	for {
		select {
		case <-ctx.Done():
			bot.StopReceivingUpdates()
			return nil
		case update := <-updates:
			if update.Message != nil {
				a.handleMessage(ctx, bot, update.Message)
			}
		}
	}
}

func redactToken(text string, token string) string {
	if token == "" {
		return text
	}
	return strings.ReplaceAll(text, token, "<redacted>")
}

type redactingTelegramLogger struct {
	token   string
	handler *channels.Handler
}

func (l redactingTelegramLogger) Println(v ...interface{}) {
	l.write(fmt.Sprintln(v...))
}

func (l redactingTelegramLogger) Printf(format string, v ...interface{}) {
	l.write(fmt.Sprintf(format, v...))
}

func (l redactingTelegramLogger) write(text string) {
	if l.handler == nil || l.handler.Log == nil {
		return
	}
	l.handler.Log.Print(strings.TrimSpace(redactToken(text, l.token)))
}

func (a *Adapter) handleMessage(ctx context.Context, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	if msg.From == nil || msg.From.IsBot {
		return
	}
	text := firstNonEmpty(msg.Text, msg.Caption)
	attachments := a.attachments(ctx, bot, msg)
	if strings.TrimSpace(text) == "" && len(attachments) == 0 {
		return
	}
	chatID := strconv.FormatInt(msg.Chat.ID, 10)
	if !channels.Allowed(chatID, a.cfg.AllowedChats) {
		return
	}
	isPrivate := msg.Chat.IsPrivate()
	botUsername := "@" + bot.Self.UserName
	mentionsBot := telegramMentionsBot(msg, bot.Self.ID, botUsername, text)
	isReplyToBot := msg.ReplyToMessage != nil && msg.ReplyToMessage.From != nil && msg.ReplyToMessage.From.ID == bot.Self.ID

	inbound := channels.InboundMessage{
		Platform:     channels.PlatformTelegram,
		MessageID:    strconv.Itoa(msg.MessageID),
		ChatID:       chatID,
		UserID:       strconv.FormatInt(msg.From.ID, 10),
		Username:     msg.From.UserName,
		IsDM:         isPrivate,
		IsGroup:      !isPrivate,
		IsReplyToBot: isReplyToBot,
		MentionsBot:  mentionsBot,
		Text:         text,
		RawText:      text,
		Attachments:  attachments,
		CreatedAt:    time.Now().UTC(),
	}

	if !channels.ShouldHandle(inbound, channels.TelegramPolicy(a.cfg)) {
		if a.handler != nil && a.handler.Log != nil {
			a.handler.Log.Printf("telegram ignored message chat=%s dm=%t mention=%t reply_to_bot=%t text_len=%d attachments=%d", chatID, isPrivate, mentionsBot, isReplyToBot, len([]rune(text)), len(attachments))
		}
		return
	}
	stopTyping := startTyping(ctx, bot, msg.Chat.ID)
	defer stopTyping()

	_ = a.handler.Handle(ctx, inbound, channels.TelegramPolicy(a.cfg), []string{botUsername}, func(sendCtx context.Context, replyTo channels.InboundMessage, reply string) error {
		return sendChunks(sendCtx, bot, msg, reply, a.handler.Text(i18n.ChannelEmptyPrompt))
	})
}

func telegramMentionsBot(msg *tgbotapi.Message, botID int64, botUsername string, text string) bool {
	botUsername = strings.ToLower(strings.TrimSpace(botUsername))
	if strings.Contains(strings.ToLower(text), botUsername) {
		return true
	}
	for _, entity := range append(msg.Entities, msg.CaptionEntities...) {
		if entity.Type == "text_mention" && entity.User != nil && entity.User.ID == botID {
			return true
		}
		if entity.Type == "mention" || entity.Type == "bot_command" {
			value := strings.ToLower(entityText(text, entity.Offset, entity.Length))
			if value == botUsername || strings.HasSuffix(value, botUsername) {
				return true
			}
		}
	}
	return false
}

func entityText(text string, offset int, length int) string {
	if offset < 0 || length <= 0 {
		return ""
	}
	encoded := utf16.Encode([]rune(text))
	if offset >= len(encoded) {
		return ""
	}
	end := offset + length
	if end > len(encoded) {
		end = len(encoded)
	}
	return string(utf16.Decode(encoded[offset:end]))
}

func startTyping(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64) func() {
	typingCtx, cancel := context.WithCancel(ctx)
	sendTyping := func() {
		_, _ = bot.Send(tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping))
	}
	sendTyping()
	go func() {
		ticker := time.NewTicker(4 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-typingCtx.Done():
				return
			case <-ticker.C:
				sendTyping()
			}
		}
	}()
	return cancel
}

func (a *Adapter) attachments(ctx context.Context, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) []channels.Attachment {
	if msg.Document == nil {
		return nil
	}
	cfg := a.attachment
	if !channels.AttachmentAllowed(msg.Document.FileName, msg.Document.MimeType, int64(msg.Document.FileSize), cfg) {
		return nil
	}
	url, err := bot.GetFileDirectURL(msg.Document.FileID)
	if err != nil {
		return nil
	}
	att, err := channels.DownloadTextAttachment(ctx, url, msg.Document.FileName, msg.Document.MimeType, int64(msg.Document.FileSize), cfg)
	if err != nil {
		return nil
	}
	return []channels.Attachment{att}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func sendChunks(ctx context.Context, bot *tgbotapi.BotAPI, replyTo *tgbotapi.Message, text, fallback string) error {
	for _, chunk := range chunks(text, fallback, 3900) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		msg := tgbotapi.NewMessage(replyTo.Chat.ID, chunk)
		msg.ParseMode = "HTML"
		msg.Text = telegramHTML(chunk)
		if !replyTo.Chat.IsPrivate() {
			msg.ReplyToMessageID = replyTo.MessageID
		}
		if _, err := bot.Send(msg); err != nil {
			return err
		}
	}
	return nil
}

var (
	telegramCodeBlockPattern = regexp.MustCompile("(?s)```(?:[a-zA-Z0-9_+-]+)?\\n?(.*?)```")
	telegramInlineCode       = regexp.MustCompile("`([^`]+)`")
	telegramBoldPattern      = regexp.MustCompile(`\*\*([^*\n][^*]*?)\*\*`)
)

func telegramHTML(text string) string {
	escaped := html.EscapeString(strings.TrimSpace(text))
	escaped = telegramCodeBlockPattern.ReplaceAllString(escaped, "<pre>$1</pre>")
	escaped = telegramInlineCode.ReplaceAllString(escaped, "<code>$1</code>")
	escaped = telegramBoldPattern.ReplaceAllString(escaped, "<b>$1</b>")
	return escaped
}

func chunks(text, fallback string, limit int) []string {
	runes := []rune(strings.TrimSpace(text))
	if len(runes) == 0 {
		return []string{fallback}
	}
	var out []string
	for len(runes) > 0 {
		n := limit
		if len(runes) < n {
			n = len(runes)
		}
		out = append(out, string(runes[:n]))
		runes = runes[n:]
	}
	return out
}
