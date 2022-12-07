package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ofen/yourcoffin/internal/bot/types"
)

const contentType = "application/json"

// Bot is bot structure.
type Bot struct {
	Token    string
	Client   *http.Client
	baseurl  string
	commands map[string]func(*types.Update) error
}

// New is bot constructor
func New(token string) *Bot {
	return &Bot{
		Token:    token,
		Client:   &http.Client{},
		baseurl:  "https://api.telegram.org/bot" + token,
		commands: make(map[string]func(*types.Update) error),
	}
}

func (b *Bot) Send(v interface{}) (*http.Response, error) {
	var method string

	switch v.(type) {
	case types.SendMessage, *types.SendMessage:
		method = "/sendMessage"
	case types.GetMyCommands, *types.GetMyCommands:
		method = "/getMyCommands"
	default:
		return nil, fmt.Errorf("not supported")
	}

	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return b.Client.Post(b.baseurl+method, contentType, bytes.NewReader(data))
}

// SendMessage sends message https://core.telegram.org/bots/api#sendmessage.
func (b *Bot) SendMessage(chatID int, text string) error {
	resp, err := b.Send(&types.SendMessage{
		Text:      MarkdownV2Escape(text),
		ChatID:    chatID,
		ParseMode: types.ParseModeMarkdownV2,
	})
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	v := &types.SendMessageResponse{}
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return err
	}

	if !v.OK {
		return fmt.Errorf("bot: %d %s", v.ErrorCode, v.Description)
	}

	return nil
}

// GetMyCommands https://core.telegram.org/bots/api#getmycommands.
func (b *Bot) GetMyCommands() ([]types.BotCommand, error) {
	resp, err := b.Send(&types.GetMyCommands{})
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	v := &types.GetMyCommandsResponse{}
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return nil, err
	}

	if !v.OK {
		return nil, fmt.Errorf("bot: %d %s", v.ErrorCode, v.Description)
	}

	return v.Result, nil
}

// Command sets bot command.
func (b *Bot) Command(command string, fn func(update *types.Update) error) {
	if !strings.HasPrefix(command, "/") {
		panic("not a command: " + command)
	}

	b.commands[command] = fn
}

// HandleCommand handles update with command.
func (b *Bot) HandleCommand(update *types.Update) error {
	if update.Message.IsBot() {
		return nil
	}

	if !update.Message.IsCommand() {
		return nil
	}

	args := update.Message.Args()
	if len(args) < 1 {
		return nil
	}

	fn, ok := b.commands[update.Message.Args()[0]]
	if !ok {
		return nil
	}

	return fn(update)
}

func MarkdownV2Escape(s string) string {
	pairs := []string{
		"_", "\\_",
		"*", "*",
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

	return strings.NewReplacer(pairs...).Replace(s)
}
