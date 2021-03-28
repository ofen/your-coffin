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
)

func getAllowedUsers() []int64 {
	allowedUsers := []int64{}

	users := strings.Split(os.Getenv("ALLOWED_USERS"), ",")
	for _, u := range users {
		id, err := strconv.ParseInt(strings.TrimSpace(u), 10, 64)
		if err == nil {
			allowedUsers = append(allowedUsers, id)
		}
	}

	return allowedUsers
}

func checkUser(fn func(bot *tgbotapi.BotAPI, update tgbotapi.Update)) func(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	return func(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
		for _, id := range allowedUsers {
			if id == update.Message.Chat.ID {
				fn(bot, update)
				return
			}
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("user not allowed: %v", update.Message.Chat.ID))
		bot.Send(msg)
		return
	}
}

func handleWhoamiCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
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

func handleStatusCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("ok"))
	bot.Send(msg)
}

func handleDefaultCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Unsupported command: %q", update.Message.Command()))
	bot.Send(msg)
}

func handleError(bot *tgbotapi.BotAPI, update tgbotapi.Update, err error) {
	log.Println(err)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, err.Error())
	bot.Send(msg)
}

func handleCurrencyCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	resp, err := http.Get("http://www.cbr.ru/scripts/XML_daily.asp")
	if err != nil {
		handleError(bot, update, err)
		return
	}
	defer resp.Body.Close()

	v := &valCurs{}
	if err = newXMLDecoder(resp.Body).Decode(&v); err != nil {
		handleError(bot, update, err)
		return
	}

	var report []string

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
	allowedUsers = getAllowedUsers()
	token := os.Getenv("BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal("unable to connect")
	}

	bot.Debug = true

	log.Printf("authorized account: %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "status":
				checkUser(handleStatusCommand)(bot, update)
			case "currency":
				checkUser(handleCurrencyCommand)(bot, update)
			case "whoami":
				handleWhoamiCommand(bot, update)
			default:
				handleDefaultCommand(bot, update)
			}

			// log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		}

	}
}
