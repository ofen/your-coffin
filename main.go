package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/text/encoding/charmap"
)

const (
	curencyEndpoint  string = "http://www.cbr.ru/scripts/XML_daily.asp"
	metersDateFormat string = "02.01.2006"
)

func newBot(token string) bot {
	api, err := tg.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatal("unable to connect")
	}

	if os.Getenv("DEBUG") == "true" {
		api.Debug = true
	}

	return bot{
		api,
		make(map[int64]handlerFunc),
	}
}

type bot struct {
	*tg.BotAPI
	nextHandler map[int64]handlerFunc
}

func (b *bot) registerNextStepHandler(message *tg.Message, f handlerFunc) {
	b.nextHandler[message.Chat.ID] = f
}

func (b *bot) nextStepHandler(message *tg.Message) handlerFunc {
	f, ok := b.nextHandler[message.Chat.ID]
	if ok {
		delete(b.nextHandler, message.Chat.ID)
		return f
	}
	return nil
}

func getAllowedUsers() []user {
	users := []user{}
	date := []byte(os.Getenv("ALLOWED_USERS"))
	json.Unmarshal(date, &users)

	return users
}

func isAllowed(message *tg.Message) bool {
	users := getAllowedUsers()

	for _, user := range users {
		if user.ID == message.Chat.ID {
			return true
		}
	}

	return false
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

func (b *bot) setCommands() error {
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

func main() {
	b := newBot(os.Getenv("BOT_TOKEN"))
	log.Printf("authorized account: %s", b.Self.UserName)

	u := tg.NewUpdate(0)
	u.Timeout = 60

	if err := b.setCommands(); err != nil {
		log.Fatal(err)
	}

	updates := b.GetUpdatesChan(u)

	for update := range updates {
		message := update.Message

		if message == nil {
			continue
		}

		if handler := b.nextStepHandler(message); handler != nil {
			handler(message)
			continue
		}

		if !message.IsCommand() {
			continue
		}

		switch isAllowed(message) {
		case true:
			switch update.Message.Command() {
			case "status":
				b.statusHandler(message)
			case "currency":
				b.currencyHandler(message)
			case "meters":
				// checkUser(handleMetersCommand)(message)
				b.metersHandler(message)
			case "whoami":
				b.whoAmIHandler(message)
			case "lastmeters":
				b.lastmetersHandler(message)
			default:
				b.unsupportedHandler(message)
			}
		case false:
			switch message.Command() {
			case "whoami":
				b.whoAmIHandler(message)
			default:
				b.unsupportedHandler(message)
			}
		}

		// log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

	}
}
