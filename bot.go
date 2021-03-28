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

func identReader(charset string, input io.Reader) (io.Reader, error) {
	if charset == "windows-1251" {
		return charmap.Windows1251.NewDecoder().Reader(input), nil
	}
	return nil, fmt.Errorf("unknown charset: %s", charset)
}

func handleWhoamiCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	text := fmt.Sprintf("*id:* %v\n*name:* %s %s", update.Message.Chat.ID, update.Message.Chat.FirstName, update.Message.Chat.LastName)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "markdown"
	bot.Send(msg)
}

func handleStatusCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("ok"))
	bot.Send(msg)
}

func handleDefaultCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Unsupported command: %q", update.Message.Command()))
	bot.Send(msg)
}

func handleCurrencyCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	var v valCurs

	resp, err := http.Get("http://www.cbr.ru/scripts/XML_daily.asp")
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	if err = xml.NewDecoder(resp.Body).Decode(&v); err != nil {
		log.Println(err)
	}

	var report []string

	for _, valute := range v.Valute {
		switch valute.CharCode {
		case
			"USD",
			"EUR":
			valuteValue, err := strconv.ParseFloat(strings.Replace(valute.Value, ",", ".", 1), 64)
			if err != nil {
				log.Fatalln(err)
			}
			report = append(report, fmt.Sprintf("*%s:* %.2f", valute.CharCode, valuteValue))
		}

	}

	messageText := strings.Join(report, "\n")

	if messageText == "" {
		messageText = "No exchange rate found"
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
	msg.ParseMode = "markdown"
	bot.Send(msg)
}

func main() {
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
				handleStatusCommand(bot, update)
			case "currency":
				handleCurrencyCommand(bot, update)
			case "whoami":
				handleWhoamiCommand(bot, update)
			default:
				handleDefaultCommand(bot, update)
			}

			// log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		}

	}
}
