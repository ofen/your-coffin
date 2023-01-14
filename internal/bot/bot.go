package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/ofen/yourcoffin/internal/bot/types"
)

// HeaderSecretToken is secret token header configured via https://core.telegram.org/bots/api#setwebhook
const HeaderSecretToken string = "x-telegram-bot-api-secret-token"

const contentType = "application/json"

// Bot is bot structure.
type Bot struct {
	Token    string
	Client   *http.Client
	baseurl  string
	commands map[string]types.HandleFunc
	next     map[int]types.HandleFunc
	mu       *sync.RWMutex
}

// New is bot constructor
func New(token string) *Bot {
	return &Bot{
		Token:    token,
		Client:   &http.Client{},
		baseurl:  "https://api.telegram.org/bot" + token,
		commands: make(map[string]types.HandleFunc),
		next:     make(map[int]types.HandleFunc),
		mu:       &sync.RWMutex{},
	}
}

func (b *Bot) getNextHandler(update *types.Update) types.HandleFunc {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.next[update.Message.Chat.ID]
}

func (b *Bot) SetNextHandler(update *types.Update, next types.HandleFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.next[update.Message.Chat.ID] = next
}

func (b *Bot) Send(method string, payload interface{}) (json.RawMessage, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("bot: %w", err)
	}

	resp, err := b.Client.Post(b.baseurl+"/"+method, contentType, bytes.NewReader(data))
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

// Command sets bot command.
func (b *Bot) Command(command string, fn types.HandleFunc) {
	if !strings.HasPrefix(command, "/") {
		panic("not a command: " + command)
	}

	b.commands[command] = fn
}

// HandleUpdate handles update with command.
func (b *Bot) HandleUpdate(ctx context.Context, u *types.Update) error {
	if u.Message.IsBot() {
		return nil
	}

	fn := b.getNextHandler(u)
	if fn != nil {
		return fn(ctx, u)
	}

	if !u.Message.IsCommand() {
		return nil
	}

	fn, ok := b.commands[u.Message.Args()[0]]
	if !ok {
		return nil
	}

	return fn(ctx, u)
}

// SendMessage sends message https://core.telegram.org/bots/api#sendmessage.
func (b *Bot) SendMessage(chatID int, text string, opts ...types.Option[types.SendMessage]) (*types.Message, error) {
	p := &types.SendMessage{
		Text:   text,
		ChatID: chatID,
	}

	for _, opt := range opts {
		opt(p)
	}

	data, err := b.Send("sendMessage", p)
	if err != nil {
		return nil, err
	}

	v := &types.Message{}
	return v, json.Unmarshal(data, v)
}

// GetMyCommands https://core.telegram.org/bots/api#getmycommands.
func (b *Bot) GetMyCommands() ([]*types.BotCommand, error) {
	data, err := b.Send("getMyCommands", types.GetMyCommands{})
	if err != nil {
		return nil, err
	}

	v := []*types.BotCommand{}
	return v, json.Unmarshal(data, &v)
}
