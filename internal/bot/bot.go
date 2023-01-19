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

func (b *Bot) Do(method string, in interface{}, out interface{}) error {
	data, err := json.Marshal(in)
	if err != nil {
		return fmt.Errorf("bot: %w", err)
	}

	resp, err := b.Client.Post(b.baseurl+"/"+method, contentType, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("bot: %w", err)
	}

	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(out)
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
func (b *Bot) SendMessage(chatID int, text string) *types.SendMessage {
	return &types.SendMessage{
		Text:   text,
		ChatID: chatID,
		Client: b,
	}
}

// GetMyCommands https://core.telegram.org/bots/api#getmycommands.
func (b *Bot) GetMyCommands() *types.GetMyCommands {
	return &types.GetMyCommands{Client: b}
}
