package bot

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

const contentType = "application/json"

// Bot is bot structure.
type Bot struct {
	Token    string
	Client   *http.Client
	baseurl  string
	commands map[string]func(*Update) error
}

// New is bot constructor
func New(token string) *Bot {
	return &Bot{
		Token:    token,
		Client:   &http.Client{},
		baseurl:  "https://api.telegram.org/bot" + token,
		commands: make(map[string]func(*Update) error),
	}
}

func (b *Bot) call(method string, req interface{}) (*http.Response, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	return b.Client.Post(b.baseurl+method, contentType, bytes.NewReader(data))
}

// SendMessage sends message https://core.telegram.org/bots/api#sendmessage.
func (b *Bot) SendMessage(chatID int, text string) error {
	resp, err := b.call("/sendMessage", &SendMessageRequest{
		Text:      strings.ReplaceAll(text, "-", "\\-"),
		ChatID:    chatID,
		ParseMode: ParseModeMarkdownV2,
	})
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)

		return errors.New(string(body))
	}

	return nil
}

// GetMyCommands https://core.telegram.org/bots/api#getmycommands.
func (b *Bot) GetMyCommands() ([]BotCommand, error) {
	resp, err := b.call("/getMyCommands", nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)

		return nil, errors.New(string(body))
	}

	result := &GetMyCommandsResponse{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}

	return result.Result, nil
}

// Command sets bot command.
func (b *Bot) Command(command string, fn func(update *Update) error) {
	b.commands[command] = fn
}

// HandleCommand handles update with command.
func (b *Bot) HandleCommand(update *Update) error {
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
