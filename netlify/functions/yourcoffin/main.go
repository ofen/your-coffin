package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/ofen/yourcoffin/internal/bot"
)

var botToken string = os.Getenv("TELEGRAM_BOT_TOKEN")

func sendMessage(text string, chatID int) *events.APIGatewayProxyResponse {
	data, err := json.Marshal(&bot.Response{
		Text:   text,
		ChatID: chatID,
	})
	if err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusServiceUnavailable,
			Body:       err.Error(),
		}
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken),
		bytes.NewReader(data))
	if err != nil {
		log.Println(err)
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusServiceUnavailable,
			Body:       err.Error(),
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusServiceUnavailable,
			Body:       err.Error(),
		}
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &events.APIGatewayProxyResponse{
			StatusCode: resp.StatusCode,
		}
	}

	return &events.APIGatewayProxyResponse{
		StatusCode: 200,
	}
}

func handler(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	log.Println(req)
	if req.HTTPMethod != http.MethodPost {
		return &events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       "Hello there!",
		}, nil
	}

	data := &bot.Data{}
	if err := json.Unmarshal([]byte(req.Body), data); err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: 503,
			Body:       err.Error(),
		}, nil
	}

	switch data.Message.Text {
	case "/status":
		return sendMessage("ok", data.Message.Chat.ID), nil
	default:
		return sendMessage("unsupported command", data.Message.Chat.ID), nil
	}
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
