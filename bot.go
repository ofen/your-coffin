package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/text/encoding/charmap"
)

const (
	curencyEndpoint  string     = "http://www.cbr.ru/scripts/XML_daily.asp"
	metersDateFormat string     = "02.01.2006"
	metersContextKey contextKey = "meters"
)

var contextKeys = []contextKey{metersContextKey}

func newBot(token string) bot {
	api, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
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
	*tgbotapi.BotAPI
	nextHandler map[int64]handlerFunc
}

func (b *bot) registerNextStepHandler(ctx context.Context, message *tgbotapi.Message, f handlerFunc) {
	b.nextHandler[message.Chat.ID] = func(c context.Context, m *tgbotapi.Message) {
		// copy context
		for _, k := range contextKeys {
			v := ctx.Value(k)
			c = context.WithValue(c, k, v)
		}
		f(c, m)
	}
}

func (b *bot) getNextStepHandler(message *tgbotapi.Message) handlerFunc {
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

func isAllowed(message *tgbotapi.Message) bool {
	users := getAllowedUsers()

	for _, user := range users {
		if user.ID == message.Chat.ID {
			return true
		}
	}

	return false
}

func (b *bot) whoAmIHandler(ctx context.Context, message *tgbotapi.Message) {
	text := fmt.Sprintf("*id:* %v\n*name:* %s %s", message.Chat.ID, message.Chat.FirstName, message.Chat.LastName)
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "markdown"
	b.Send(msg)
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

func (b *bot) statusHandler(ctx context.Context, message *tgbotapi.Message) {
	text := ""
	version := os.Getenv("SOURCE_VERSION")

	switch {
	case version != "":
		text = fmt.Sprintf("version: %s", version)
	default:
		text = "ok"
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	b.Send(msg)
}

func (b *bot) metersHandler(ctx context.Context, message *tgbotapi.Message) {
	b.Send(tgbotapi.NewMessage(message.Chat.ID, "enter hot water or use /cancel to stop"))
	ctx = context.WithValue(ctx, metersContextKey, &meters{-1, -1, -1, -1})
	b.registerNextStepHandler(ctx, message, b.handleMeters)
}

func (b *bot) handleMeters(ctx context.Context, message *tgbotapi.Message) {
	if message.Text == "/cancel" {
		b.Send(tgbotapi.NewMessage(message.Chat.ID, "canceled"))
		return
	}
	v, err := strconv.Atoi(message.Text)
	if err != nil || v < 0 {
		b.Send(tgbotapi.NewMessage(message.Chat.ID, "value should be positive number"))
		b.registerNextStepHandler(ctx, message, b.handleMeters)
		return
	}

	m := ctx.Value(metersContextKey).(*meters)

	switch {
	case m.hotWater < 0:
		m.setHotWater(v)
		b.Send(tgbotapi.NewMessage(message.Chat.ID, "enter cold water"))
		b.registerNextStepHandler(ctx, message, b.handleMeters)
		return
	case m.coldWater < 0:
		m.setColdWater(v)
		b.Send(tgbotapi.NewMessage(message.Chat.ID, "enter electricity (t1)"))
		b.registerNextStepHandler(ctx, message, b.handleMeters)
		return
	case m.electricityT1 < 0:
		m.setElectricityT1(v)
		b.Send(tgbotapi.NewMessage(message.Chat.ID, "enter electricity (t2)"))
		b.registerNextStepHandler(ctx, message, b.handleMeters)
		return
	case m.electricityT2 < 0:
		m.setElectricityT2(v)
		text := fmt.Sprintf("\n*hot water:* %v", m.hotWater)
		text += fmt.Sprintf("\n*cold water:* %v", m.coldWater)
		text += fmt.Sprintf("\n*electricity (t1):* %v", m.electricityT1)
		text += fmt.Sprintf("\n*electricity (t2):* %v", m.electricityT2)
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		msg.ParseMode = tgbotapi.ModeMarkdown
		b.Send(msg)
	}

	s, err := newSpreadsheet(os.Getenv("GOOGLE_SPREADSHEET"))
	if err != nil {
		b.Send(tgbotapi.NewMessage(message.Chat.ID, err.Error()))
		return
	}

	today := message.Time().Format(metersDateFormat)
	err = s.appendRow([]interface{}{
		today,
		m.hotWater,
		m.coldWater,
		m.electricityT1,
		m.electricityT2,
	})
	if err != nil {
		b.Send(tgbotapi.NewMessage(message.Chat.ID, err.Error()))
		return
	}
	b.Send(tgbotapi.NewMessage(message.Chat.ID, "sheet updated"))
}

func (b *bot) unsupportedHandler(ctx context.Context, message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("unsupported command: %q", message.Command()))
	b.Send(msg)
}

func (b *bot) handleError(message *tgbotapi.Message, err error) {
	log.Println(err)
	msg := tgbotapi.NewMessage(message.Chat.ID, err.Error())
	b.Send(msg)
}

func (b *bot) currencyHandler(ctx context.Context, message *tgbotapi.Message) {
	resp, err := http.Get(curencyEndpoint)
	if err != nil {
		b.handleError(message, err)
		return
	}
	defer resp.Body.Close()

	v := &valCurs{}
	if err = newXMLDecoder(resp.Body).Decode(&v); err != nil {
		b.handleError(message, err)
		return
	}

	report := []string{}

	for _, valute := range v.Valute {
		switch strings.ToLower(valute.CharCode) {
		case "usd", "eur":
			valuteValue, err := strconv.ParseFloat(strings.Replace(valute.Value, ",", ".", 1), 64)
			if err != nil {
				log.Fatalln(err)
			}
			report = append(report, fmt.Sprintf("*%s:* %.2f", valute.CharCode, valuteValue))
		}

	}

	text := strings.Join(report, "\n")

	if text == "" {
		text = "No exchange rate found"
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "markdown"
	b.Send(msg)
}

func main() {
	b := newBot(os.Getenv("BOT_TOKEN"))
	log.Printf("authorized account: %s", b.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}

	for update := range updates {
		message := update.Message
		ctx := context.TODO()

		if message == nil {
			continue
		}

		if handler := b.getNextStepHandler(message); handler != nil {
			handler(ctx, message)
			continue
		}

		if !message.IsCommand() {
			continue
		}

		switch isAllowed(message) {
		case true:
			switch update.Message.Command() {
			case "status":
				b.statusHandler(ctx, message)
			case "currency":
				b.currencyHandler(ctx, message)
			case "meters":
				// checkUser(handleMetersCommand)(message)
				b.metersHandler(ctx, message)
			case "whoami":
				b.whoAmIHandler(ctx, message)
			default:
				b.unsupportedHandler(ctx, message)
			}
		case false:
			switch message.Command() {
			case "whoami":
				b.whoAmIHandler(ctx, message)
			default:
				b.unsupportedHandler(ctx, message)
			}
		}

		// log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

	}
}
