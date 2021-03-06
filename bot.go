package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/text/encoding/charmap"
)

var (
	allowedUsers []int64
	bot          *tgbotapi.BotAPI
	sheet        *spreadsheet
)

func getAllowedUsers(usersList string) []int64 {
	allowedUsers := []int64{}

	users := strings.Split(usersList, ",")
	for _, u := range users {
		id, err := strconv.ParseInt(strings.TrimSpace(u), 10, 64)
		if err == nil {
			allowedUsers = append(allowedUsers, id)
		}
	}

	return allowedUsers
}

func checkUser(fn func(update tgbotapi.Update)) func(update tgbotapi.Update) {
	return func(update tgbotapi.Update) {

		if len(allowedUsers) == 0 {
			fn(update)
			return
		}

		for _, id := range allowedUsers {
			if id == update.Message.Chat.ID {
				fn(update)
				return
			}
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("user not allowed: %v", update.Message.Chat.ID))
		bot.Send(msg)
		return
	}
}

func handleWhoamiCommand(update tgbotapi.Update) {
	text := fmt.Sprintf("*id:* %v\n*name:* %s %s", update.Message.Chat.ID, update.Message.Chat.FirstName, update.Message.Chat.LastName)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "markdown"
	bot.Send(msg)
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

func handleStatusCommand(update tgbotapi.Update) {
	text := ""
	version := os.Getenv("SOURCE_VERSION")

	switch {
	case version != "":
		text = fmt.Sprintf("version: %s", version)
	default:
		text = "ok"
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	bot.Send(msg)
}

func handleMetersCommand(update tgbotapi.Update) {
	date := update.Message.Time().Format("02-01-2006")
	args := []int{}

	for _, arg := range strings.Split(update.Message.CommandArguments(), " ") {
		int, err := strconv.Atoi(arg)
		if err == nil {
			args = append(args, int)
		}
	}

	if len(args) != 4 {
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "incorrect arguments"))
		return
	}

	sheet.appendRow([]interface{}{date, args[0], args[1], args[2], args[3]})

	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprint("sheet updated")))
}

func handleUnsupportedCommand(update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Unsupported command: %q", update.Message.Command()))
	bot.Send(msg)
}

func handleError(update tgbotapi.Update, err error) {
	log.Println(err)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, err.Error())
	bot.Send(msg)
}

func handleCurrencyCommand(update tgbotapi.Update) {
	resp, err := http.Get("http://www.cbr.ru/scripts/XML_daily.asp")
	if err != nil {
		handleError(update, err)
		return
	}
	defer resp.Body.Close()

	v := &valCurs{}
	if err = newXMLDecoder(resp.Body).Decode(&v); err != nil {
		handleError(update, err)
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

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "markdown"
	bot.Send(msg)
}

func main() {
	s, err := newSpreadsheet(os.Getenv("GOOGLE_SPREADSHEET"))
	if err == nil {
		sheet = s
	}
	allowedUsers = getAllowedUsers(os.Getenv("ALLOWED_USERS"))
	b, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatal("unable to connect")
	}

	bot = b
	bot.Debug = true

	log.Printf("authorized account: %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "status":
				checkUser(handleStatusCommand)(update)
			case "currency":
				checkUser(handleCurrencyCommand)(update)
			case "meters":
				checkUser(handleMetersCommand)(update)
			case "whoami":
				handleWhoamiCommand(update)
			default:
				handleUnsupportedCommand(update)
			}

			// log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		}

	}
}
