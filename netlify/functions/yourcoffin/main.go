package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const contentType = "application/json"

type Update struct {
	UpdateID int     `json:"update_id"`
	Message  Message `json:"message"`
}

type Response struct {
	Text   string `json:"text"`
	ChatID int    `json:"chat_id"`
}

type Message struct {
	MessageID int    `json:"message_id"`
	From      From   `json:"from"`
	Chat      Chat   `json:"chat"`
	Date      int    `json:"date"`
	Text      string `json:"text"`
}

func (m *Message) IsCommand() bool {
	return strings.HasPrefix(m.Text, "/")
}

type Chat struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
	Type      string `json:"type"`
}

type From struct {
	ID           int    `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
}

var botToken string = os.Getenv("BOT_TOKEN")

type Bot struct {
	token   string
	baseurl string
}

func New(token string) *Bot {
	return &Bot{
		token:   token,
		baseurl: "https://api.telegram.org/bot" + token,
	}
}

func (b *Bot) SendMessage(chatID int, text string) error {
	data, err := json.Marshal(&Response{
		Text:   text,
		ChatID: chatID,
	})
	if err != nil {
		return err
	}

	resp, err := http.Post(b.baseurl+"/sendMessage", contentType, bytes.NewReader(data))
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

func process(data []byte) error {
	update := &Update{}
	if err := json.Unmarshal(data, update); err != nil {
		return err
	}

	if !update.Message.IsCommand() {
		return nil
	}

	bot := New(botToken)

	switch update.Message.Text {
	case "/status":
		return bot.SendMessage(update.Message.Chat.ID, "ok")
	default:
		return bot.SendMessage(update.Message.Chat.ID, "unsupported command")
	}
}

func handler(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	log.Println(req)

	if req.HTTPMethod != http.MethodPost {
		return &events.APIGatewayProxyResponse{StatusCode: 403}, nil
	}

	if err := process([]byte(req.Body)); err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: 503,
			Body:       err.Error(),
		}, nil
	}

	return &events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

func main() {
	lambda.Start(handler)
}

// import (
// 	"log"
// 	"os"

// 	"github.com/ofen/yourcoffin/internal/bot"
// )

// func main() {
// 	b, err := bot.New(os.Getenv("BOT_TOKEN"))
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	log.Printf("authorized account: %s", b.Self.UserName)

// 	for update := range b.HandleUpdate() {
// 		message := update.Message

// 		if message == nil {
// 			continue
// 		}

// 		if handler := b.NextStepHandler(message); handler != nil {
// 			handler(message)
// 			continue
// 		}

// 		if !message.IsCommand() {
// 			continue
// 		}

// 		switch bot.IsAllowed(message) {
// 		case true:
// 			switch update.Message.Command() {
// 			case "status":
// 				b.StatusHandler(message)
// 			case "currency":
// 				b.CurrencyHandler(message)
// 			case "meters":
// 				// checkUser(handleMetersCommand)(message)
// 				b.MetersHandler(message)
// 			case "whoami":
// 				b.WhoAmIHandler(message)
// 			case "lastmeters":
// 				b.LastmetersHandler(message)
// 			default:
// 				b.UnsupportedHandler(message)
// 			}
// 		case false:
// 			switch message.Command() {
// 			case "whoami":
// 				b.WhoAmIHandler(message)
// 			default:
// 				b.UnsupportedHandler(message)
// 			}
// 		}

// 		// log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

// 	}
// }
