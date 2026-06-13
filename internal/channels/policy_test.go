package channels

import (
	"testing"
	"vietclaw/internal/config"
)

func TestShouldHandle(t *testing.T) {
	tests := []struct {
		name   string
		msg    InboundMessage
		policy Policy
		want   bool
	}{
		// DM tests
		{
			name:   "DM allowed",
			msg:    InboundMessage{IsDM: true},
			policy: Policy{RespondInDM: true},
			want:   true,
		},
		{
			name:   "DM denied",
			msg:    InboundMessage{IsDM: true},
			policy: Policy{RespondInDM: false},
			want:   false,
		},

		// Group 'always' policy
		{
			name:   "Group always - no mention",
			msg:    InboundMessage{IsDM: false, MentionsBot: false, IsReplyToBot: false},
			policy: Policy{RespondInGroups: "always"},
			want:   true,
		},
		{
			name:   "Group ALWAYS (case insensitive)",
			msg:    InboundMessage{IsDM: false},
			policy: Policy{RespondInGroups: "ALWAYS"},
			want:   true,
		},
		{
			name:   "Group always with spaces",
			msg:    InboundMessage{IsDM: false},
			policy: Policy{RespondInGroups: " always "},
			want:   true,
		},

		// Group 'never/off/false' policy
		{
			name:   "Group never",
			msg:    InboundMessage{IsDM: false, MentionsBot: true}, // Even if mentioned
			policy: Policy{RespondInGroups: "never"},
			want:   false,
		},
		{
			name:   "Group off",
			msg:    InboundMessage{IsDM: false, IsReplyToBot: true},
			policy: Policy{RespondInGroups: "off"},
			want:   false,
		},
		{
			name:   "Group false",
			msg:    InboundMessage{IsDM: false},
			policy: Policy{RespondInGroups: "false"},
			want:   false,
		},

		// Group default policy (mention or reply)
		{
			name:   "Group default - mentions bot",
			msg:    InboundMessage{IsDM: false, MentionsBot: true, IsReplyToBot: false},
			policy: Policy{RespondInGroups: "mentions"}, // Any string not 'always' or 'never/off/false'
			want:   true,
		},
		{
			name:   "Group default - replies to bot",
			msg:    InboundMessage{IsDM: false, MentionsBot: false, IsReplyToBot: true},
			policy: Policy{RespondInGroups: ""}, // Empty string falls to default
			want:   true,
		},
		{
			name:   "Group default - no mention or reply",
			msg:    InboundMessage{IsDM: false, MentionsBot: false, IsReplyToBot: false},
			policy: Policy{RespondInGroups: "default"},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldHandle(tt.msg, tt.policy); got != tt.want {
				t.Errorf("ShouldHandle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStripMentions(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		mentions []string
		want     string
	}{
		{
			name:     "No mentions",
			text:     "hello world",
			mentions: []string{"@bot"},
			want:     "hello world",
		},
		{
			name:     "Single mention at start",
			text:     "@bot hello",
			mentions: []string{"@bot"},
			want:     "hello",
		},
		{
			name:     "Mention with different case",
			text:     "@Bot hello",
			mentions: []string{"@bot"},
			want:     "hello",
		},
		{
			name:     "Multiple different mentions",
			text:     "@bot1 @bot2 hello",
			mentions: []string{"@bot1", "@bot2"},
			want:     "hello",
		},
		{
			name:     "Mention not at start",
			text:     "hello @bot",
			mentions: []string{"@bot"},
			want:     "hello @bot",
		},
		{
			name:     "Empty mentions list",
			text:     "@bot hello",
			mentions: []string{},
			want:     "@bot hello",
		},
		{
			name:     "Mentions list with empty string",
			text:     "@bot hello",
			mentions: []string{"", "@bot"},
			want:     "hello",
		},
		{
			name:     "Whitespace handling",
			text:     "  @bot   hello  ",
			mentions: []string{"@bot"},
			want:     "hello",
		},
		{
			name:     "Consecutive identical mentions",
			text:     "@bot @bot hello",
			mentions: []string{"@bot"},
			want:     "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripMentions(tt.text, tt.mentions); got != tt.want {
				t.Errorf("StripMentions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscordPolicy(t *testing.T) {
	cfg := config.DiscordConfig{
		RespondInDM:     true,
		RespondInGuilds: "always",
	}
	policy := DiscordPolicy(cfg)

	if !policy.RespondInDM {
		t.Errorf("Expected RespondInDM to be true")
	}
	if policy.RespondInGroups != "always" {
		t.Errorf("Expected RespondInGroups to be 'always', got '%s'", policy.RespondInGroups)
	}
}

func TestTelegramPolicy(t *testing.T) {
	cfg := config.TelegramConfig{
		RespondInPrivate: true,
		RespondInGroups:  "never",
	}
	policy := TelegramPolicy(cfg)

	if !policy.RespondInDM {
		t.Errorf("Expected RespondInDM to be true")
	}
	if policy.RespondInGroups != "never" {
		t.Errorf("Expected RespondInGroups to be 'never', got '%s'", policy.RespondInGroups)
	}
}

func TestAllowed(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		allowed []string
		want    bool
	}{
		{
			name:    "Empty allowed list",
			value:   "anything",
			allowed: []string{},
			want:    true,
		},
		{
			name:    "Nil allowed list",
			value:   "anything",
			allowed: nil,
			want:    true,
		},
		{
			name:    "Value in allowed list",
			value:   "allowed_val",
			allowed: []string{"other", "allowed_val", "another"},
			want:    true,
		},
		{
			name:    "Value not in allowed list",
			value:   "denied_val",
			allowed: []string{"other", "allowed_val", "another"},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Allowed(tt.value, tt.allowed); got != tt.want {
				t.Errorf("Allowed() = %v, want %v", got, tt.want)
			}
		})
	}
}
