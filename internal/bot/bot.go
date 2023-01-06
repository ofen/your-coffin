package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ofen/yourcoffin/internal/bot/types"
)

// HeaderSecretToken is secret token header configured via https://core.telegram.org/bots/api#setwebhook
const HeaderSecretToken string = "x-telegram-bot-api-secret-token"

const contentType = "application/json"

type Request interface {
	Method() string
}

// Bot is bot structure.
type Bot struct {
	Token    string
	Client   *http.Client
	baseurl  string
	commands map[string]types.HandleFunc
}

// New is bot constructor
func New(token string) *Bot {
	return &Bot{
		Token:    token,
		Client:   &http.Client{},
		baseurl:  "https://api.telegram.org/bot" + token,
		commands: make(map[string]types.HandleFunc),
	}
}

func (b *Bot) Send(r Request) (json.RawMessage, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("bot: %w", err)
	}

	resp, err := b.Client.Post(b.baseurl+"/"+r.Method(), contentType, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("bot: %w", err)
	}

	defer resp.Body.Close()

	v := &types.Response[json.RawMessage]{}
	if err = json.NewDecoder(resp.Body).Decode(v); err != nil {
		return nil, err
	}

	if err = v.IsError(); err != nil {
		return nil, err
	}

	return v.Result, nil
}

// SendMessage sends message https://core.telegram.org/bots/api#sendmessage.
func (b *Bot) SendMessage(chatID int, text string, parseMode types.ParseMode) (*types.Message, error) {
	data, err := b.Send(types.SendMessage{
		Text:      text,
		ChatID:    chatID,
		ParseMode: parseMode,
	})
	if err != nil {
		return nil, err
	}

	v := &types.Message{}
	return v, json.Unmarshal(data, v)
}

// GetMyCommands https://core.telegram.org/bots/api#getmycommands.
func (b *Bot) GetMyCommands() ([]*types.BotCommand, error) {
	data, err := b.Send(types.GetMyCommands{})
	if err != nil {
		return nil, err
	}

	v := []*types.BotCommand{}
	return v, json.Unmarshal(data, &v)
}

// Command sets bot command.
func (b *Bot) Command(command string, fn types.HandleFunc) {
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
