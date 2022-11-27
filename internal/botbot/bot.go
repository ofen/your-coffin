package botbot

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/text/encoding/charmap"
)

const (
	curencyEndpoint  string = "http://www.cbr.ru/scripts/XML_daily.asp"
	metersDateFormat string = "02.01.2006"
)

func New(token string) (*Bot, error) {
	api, err := tg.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		return nil, err
	}

	if os.Getenv("DEBUG") == "true" {
		api.Debug = true
	}

	return &Bot{
		api,
		make(map[int64]handlerFunc),
	}, nil
}

type Bot struct {
	*tg.BotAPI
	nextHandler map[int64]handlerFunc
}

func (b *Bot) RegisterNextStepHandler(message *tg.Message, f handlerFunc) {
	b.nextHandler[message.Chat.ID] = f
}

func (b *Bot) NextStepHandler(message *tg.Message) handlerFunc {
	f, ok := b.nextHandler[message.Chat.ID]
	if ok {
		delete(b.nextHandler, message.Chat.ID)
		return f
	}
	return nil
}

func newXMLDecoder(r io.Reader) *xml.Decoder {
	d := xml.NewDecoder(r)

	d.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		if charset == "windows-1251" {
			return charmap.Windows1251.NewDecoder().Reader(input), nil
		}
		return nil, fmt.Errorf("unknown charset: %s", charset)
	}

	return d
}

func (b *Bot) SetCommands() error {
	commands := tg.NewSetMyCommands(
		tg.BotCommand{Command: "status", Description: "check bot status"},
		tg.BotCommand{Command: "currency", Description: "chech current exchange rate"},
		tg.BotCommand{Command: "meters", Description: "set meters"},
		tg.BotCommand{Command: "lastmeters", Description: "show last meters"},
		tg.BotCommand{Command: "whoami", Description: "show info about requesting user"},
	)
	r, err := b.Request(commands)
	if err != nil {
		return err
	}

	if !r.Ok {
		return fmt.Errorf(string(r.Result))
	}

	return nil
}
