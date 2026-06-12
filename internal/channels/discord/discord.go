package discord

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"vietclaw/internal/channels"
	"vietclaw/internal/config"
	"vietclaw/internal/i18n"
)

type Adapter struct {
	cfg        config.DiscordConfig
	attachment config.AttachmentConfig
	handler    *channels.Handler
}

func init() {
	channels.RegisterAdapter(channels.PlatformDiscord, func(cfg config.Config, handler *channels.Handler) (channels.Adapter, error) {
		return New(cfg.Channels.Discord, cfg.Channels.Attachments, handler), nil
	})
}

func New(cfg config.DiscordConfig, attachment config.AttachmentConfig, handler *channels.Handler) *Adapter {
	return &Adapter{cfg: cfg, attachment: attachment, handler: handler}
}

func (a *Adapter) Name() string {
	return channels.PlatformDiscord
}

func (a *Adapter) Start(ctx context.Context) error {
	token := strings.TrimSpace(os.Getenv(a.cfg.TokenEnv))
	if token == "" {
		return fmt.Errorf("discord token env missing: %s", a.cfg.TokenEnv)
	}
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return fmt.Errorf("create discord session: %w", err)
	}
	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages | discordgo.IntentMessageContent | discordgo.IntentGuilds

	session.AddHandler(a.onMessage(ctx))
	session.AddHandler(a.onInteraction(ctx))
	if err := session.Open(); err != nil {
		return fmt.Errorf("open discord session: %w", err)
	}
	if session.State != nil && session.State.User != nil {
		_, _ = session.ApplicationCommandCreate(session.State.User.ID, "", &discordgo.ApplicationCommand{
			Name:        "models",
			Description: "List or select the AI model catalog entry",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "id",
					Description: "Catalog model id to select",
					Required:    false,
				},
			},
		})
	}
	defer session.Close()

	<-ctx.Done()
	return nil
}

func (a *Adapter) onInteraction(ctx context.Context) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, event *discordgo.InteractionCreate) {
		if event == nil || event.ApplicationCommandData().Name != "models" {
			return
		}
		userID := ""
		if event.Member != nil && event.Member.User != nil {
			userID = event.Member.User.ID
		} else if event.User != nil {
			userID = event.User.ID
		}
		arg := ""
		for _, opt := range event.ApplicationCommandData().Options {
			if opt.Name == "id" {
				arg = opt.StringValue()
			}
		}
		text := "/models"
		if arg != "" {
			text += " " + arg
		}
		result, err := channels.HandleModelCommand(ctx, a.handler.DB, a.handler.Config, channels.PlatformDiscord, userID, text)
		if err != nil {
			_ = s.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "Failed to handle /models"},
			})
			return
		}
		_ = s.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: result.Reply},
		})
	}
}

func (a *Adapter) onMessage(ctx context.Context) func(*discordgo.Session, *discordgo.MessageCreate) {
	return func(s *discordgo.Session, event *discordgo.MessageCreate) {
		msg := event.Message
		if msg == nil || msg.Author == nil || msg.Author.Bot {
			return
		}
		attachments := a.attachments(ctx, msg.Attachments)
		if strings.TrimSpace(msg.Content) == "" && len(attachments) == 0 {
			return
		}
		if s.State != nil && s.State.User != nil && msg.Author.ID == s.State.User.ID {
			return
		}
		if msg.GuildID != "" && !channels.Allowed(msg.GuildID, a.cfg.AllowedGuilds) {
			return
		}
		if !channels.Allowed(msg.ChannelID, a.cfg.AllowedChannels) {
			return
		}

		botID := ""
		if s.State != nil && s.State.User != nil {
			botID = s.State.User.ID
		}
		mentionsBot := mentionsUser(msg, botID)
		isReplyToBot := msg.ReferencedMessage != nil && msg.ReferencedMessage.Author != nil && msg.ReferencedMessage.Author.ID == botID
		isDM := msg.GuildID == ""
		threadID := ""
		if msg.Thread != nil {
			threadID = msg.Thread.ID
		}

		inbound := channels.InboundMessage{
			Platform:     channels.PlatformDiscord,
			MessageID:    msg.ID,
			GuildID:      msg.GuildID,
			ChannelID:    msg.ChannelID,
			ThreadID:     threadID,
			UserID:       msg.Author.ID,
			Username:     msg.Author.Username,
			IsDM:         isDM,
			IsGroup:      !isDM,
			IsReplyToBot: isReplyToBot,
			MentionsBot:  mentionsBot,
			Text:         msg.Content,
			RawText:      msg.Content,
			Attachments:  attachments,
			CreatedAt:    time.Now().UTC(),
		}

		if !channels.ShouldHandle(inbound, channels.DiscordPolicy(a.cfg)) {
			return
		}
		_ = s.ChannelTyping(msg.ChannelID)
		mentions := []string{"<@" + botID + ">", "<@!" + botID + ">"}
		_ = a.handler.Handle(ctx, inbound, channels.DiscordPolicy(a.cfg), mentions, func(sendCtx context.Context, replyTo channels.InboundMessage, reply string) error {
			fallback := a.handler.Text(i18n.ChannelEmptyPrompt)
			return sendChunks(sendCtx, s, replyTo.ChannelID, replyTo.MessageID, reply, fallback)
		})
	}
}

func (a *Adapter) attachments(ctx context.Context, items []*discordgo.MessageAttachment) []channels.Attachment {
	cfg := a.attachment
	maxFiles := cfg.MaxFiles
	if maxFiles <= 0 {
		maxFiles = config.DefaultAttachmentMaxFiles
	}
	var out []channels.Attachment
	for _, item := range items {
		if item == nil || len(out) >= maxFiles {
			continue
		}
		if !channels.AttachmentAllowed(item.Filename, item.ContentType, int64(item.Size), cfg) {
			continue
		}
		att, err := channels.DownloadTextAttachment(ctx, item.URL, item.Filename, item.ContentType, int64(item.Size), cfg)
		if err == nil {
			out = append(out, att)
		}
	}
	return out
}

func mentionsUser(msg *discordgo.Message, userID string) bool {
	if userID == "" {
		return false
	}
	for _, mention := range msg.Mentions {
		if mention != nil && mention.ID == userID {
			return true
		}
	}
	return strings.Contains(msg.Content, "<@"+userID+">") || strings.Contains(msg.Content, "<@!"+userID+">")
}

func sendChunks(ctx context.Context, session *discordgo.Session, channelID, referenceID, text, fallback string) error {
	for _, chunk := range chunks(text, fallback, 1900) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		message := &discordgo.MessageSend{Content: chunk}
		if referenceID != "" {
			message.Reference = &discordgo.MessageReference{MessageID: referenceID, ChannelID: channelID}
		}
		if _, err := session.ChannelMessageSendComplex(channelID, message); err != nil {
			return err
		}
	}
	return nil
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
