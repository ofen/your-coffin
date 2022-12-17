package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// MessageEntity types https://core.telegram.org/bots/api#messageentity.
const (
	MessageEntityBotCommand    string = "bot_command"
	MessageEntityMention       string = "mention"
	MessageEntityHashTag       string = "hashtag"
	MessageEntityCashTag       string = "cashtag"
	MessageEntityURL           string = "url"
	MessageEntityEmail         string = "email"
	MessageEntityPhoneNumber   string = "phone_number"
	MessageEntityBold          string = "bold"
	MessageEntityItalic        string = "italic"
	MessageEntityUnderline     string = "underline"
	MessageEntityStrikethrough string = "strikethrough"
	MessageEntitySpoiler       string = "spoiler"
	MessageEntityCode          string = "code"
	MessageEntityPre           string = "pre"
	MessageEntityTextLink      string = "text_link"
	MessageEntityTextMention   string = "text_mention"
	MessageEntityCustomEmoji   string = "custom_emoji"
)

// https://core.telegram.org/bots/api#formatting-options.
const (
	ParseModeMarkdownV2 string = "MarkdownV2"
	ParseModeHTML       string = "HTML"
	ParseModeMarkdown   string = "Markdown"
)

type SendMessageOption struct {
	ParseMode string
}

// SendMessage https://core.telegram.org/bots/api#sendmessage.
type SendMessage struct {
	Text      string  `json:"text"`
	ChatID    int     `json:"chat_id"`
	ParseMode string  `json:"parse_mode"`
	Result    Message `json:"result"`
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

func (t *Message) UnmarshalJSON(data []byte) error {
	type alias Message

	tmp := struct {
		Date int
		*alias
	}{
		alias: (*alias)(t),
	}

	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	t.Date = time.Unix(int64(tmp.Date), 0)

	return nil
}

func (m *Message) IsBot() bool {
	return m.From.IsBot
}

func (m *Message) IsCommand() bool {
	for _, e := range m.Entities {
		if e.Type == MessageEntityBotCommand {
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
	Offeset       int    `json:"offset"`
	Length        int    `json:"length"`
	Type          string `json:"type"`
	URL           string `json:"url"`
	User          User   `json:"user"`
	Language      string `json:"language"`
	CustomEmojiID string `json:"custom_emoji_id"`
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
