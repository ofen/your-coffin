package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type HandleFunc func(update *Update) error

type MessageEntityType string

// MessageEntity types https://core.telegram.org/bots/api#messageentity.
const (
	MessageEntityTypeBotCommand    MessageEntityType = "bot_command"
	MessageEntityTypeMention       MessageEntityType = "mention"
	MessageEntityTypeHashTag       MessageEntityType = "hashtag"
	MessageEntityTypeCashTag       MessageEntityType = "cashtag"
	MessageEntityTypeURL           MessageEntityType = "url"
	MessageEntityTypeEmail         MessageEntityType = "email"
	MessageEntityTypePhoneNumber   MessageEntityType = "phone_number"
	MessageEntityTypeBold          MessageEntityType = "bold"
	MessageEntityTypeItalic        MessageEntityType = "italic"
	MessageEntityTypeUnderline     MessageEntityType = "underline"
	MessageEntityTypeStrikethrough MessageEntityType = "strikethrough"
	MessageEntityTypeSpoiler       MessageEntityType = "spoiler"
	MessageEntityTypeCode          MessageEntityType = "code"
	MessageEntityTypePre           MessageEntityType = "pre"
	MessageEntityTypeTextLink      MessageEntityType = "text_link"
	MessageEntityTypeTextMention   MessageEntityType = "text_mention"
	MessageEntityTypeCustomEmoji   MessageEntityType = "custom_emoji"
)

type ParseMode string

// https://core.telegram.org/bots/api#formatting-options.
const (
	ParseModeMarkdownV2 ParseMode = "MarkdownV2"
	ParseModeHTML       ParseMode = "HTML"
	ParseModeMarkdown   ParseMode = "Markdown"
)

// SendMessage https://core.telegram.org/bots/api#sendmessage.
type SendMessage struct {
	Text      string    `json:"text"`
	ChatID    int       `json:"chat_id"`
	ParseMode ParseMode `json:"parse_mode"`
}

func (m SendMessage) MarshalJSON() ([]byte, error) {
	type Alias SendMessage

	v := (Alias)(m)
	v.Text = escapeText(v.ParseMode, v.Text)

	return json.Marshal(v)
}

func (m SendMessage) Method() string {
	return "sendMessage"
}

type SendMessageResponse struct {
	Response[Message]
}

type GetMyCommands struct{}

func (m GetMyCommands) Method() string {
	return "getMyCommands"
}

// GetMyCommandsResponse https://core.telegram.org/bots/api#getmycommands.
type GetMyCommandsResponse struct {
	Response[[]BotCommand]
}

type BotCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

// Update https://core.telegram.org/bots/api#update.
type Update struct {
	UpdateID int     `json:"update_id"`
	Message  Message `json:"message"`
}

// Message https://core.telegram.org/bots/api#message.
type Message struct {
	MessageID int             `json:"message_id"`
	From      User            `json:"from"`
	Chat      Chat            `json:"chat"`
	Date      time.Time       `json:"date"`
	Text      string          `json:"text"`
	Entities  []MessageEntity `json:"entities"`
}

func (m *Message) UnmarshalJSON(data []byte) error {
	type Alias Message

	tmp := struct {
		Date int `json:"date"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	m.Date = time.Unix(int64(tmp.Date), 0)

	return nil
}

func (m *Message) IsBot() bool {
	return m.From.IsBot
}

func (m *Message) IsCommand() bool {
	for _, e := range m.Entities {
		if e.Type == MessageEntityTypeBotCommand {
			return true
		}
	}

	return false
}

func (m *Message) Args() []string {
	if !m.IsCommand() {
		return []string{}
	}

	return strings.Fields(m.Text)
}

// Chat https://core.telegram.org/bots/api#chat.
type Chat struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
	Type      string `json:"type"`
}

// User https://core.telegram.org/bots/api#user.
type User struct {
	ID           int    `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
}

// MessageEntity https://core.telegram.org/bots/api#messageentity.
type MessageEntity struct {
	Offeset       int               `json:"offset"`
	Length        int               `json:"length"`
	Type          MessageEntityType `json:"type"`
	URL           string            `json:"url"`
	User          User              `json:"user"`
	Language      string            `json:"language"`
	CustomEmojiID string            `json:"custom_emoji_id"`
}

type Response[T any] struct {
	OK          bool   `json:"ok"`
	ErrorCode   int    `json:"error_code"`
	Description string `json:"description"`
	Result      T      `json:"result"`
}

func (r Response[T]) IsError() error {
	if r.OK {
		return nil
	}

	return fmt.Errorf("%d %s", r.ErrorCode, r.Description)
}

func escapeText(parseMode ParseMode, text string) string {
	var pairs []string

	switch parseMode {
	case ParseModeHTML:
		pairs = []string{
			"<", "&lt;",
			">", "&gt;",
			"&", "&amp;",
		}
	case ParseModeMarkdown:
		pairs = []string{
			"_", "\\_",
			"*", "\\*",
			"`", "\\`",
			"[", "\\[",
		}
	case ParseModeMarkdownV2:
		pairs = []string{
			"_", "\\_",
			"[", "\\[",
			"]", "\\]",
			"(", "\\(",
			")", "\\)",
			"~", "\\~",
			"`", "\\`",
			">", "\\>",
			"#", "\\#",
			"+", "\\+",
			"-", "\\-",
			"=", "\\=",
			"|", "\\|",
			"{", "\\{",
			"}", "\\}",
			".", "\\.",
			"!", "\\!",
		}
	default:
		return text
	}

	return strings.NewReplacer(pairs...).Replace(text)
}
