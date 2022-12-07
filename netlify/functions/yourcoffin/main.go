package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/ofen/yourcoffin/internal/bot"
	"github.com/ofen/yourcoffin/internal/bot/types"
)

var b = bot.New(os.Getenv("BOT_TOKEN"))

func handler(r events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	log.Println(r)

	if r.HTTPMethod != http.MethodPost {
		return &events.APIGatewayProxyResponse{StatusCode: http.StatusForbidden}, nil
	}

	update := &types.Update{}
	if err := json.Unmarshal([]byte(r.Body), update); err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusUnprocessableEntity,
			Body:       err.Error(),
		}, nil
	}

	if err := b.HandleCommand(update); err != nil {
		log.Println(err)
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, nil
	}

	return &events.APIGatewayProxyResponse{StatusCode: http.StatusOK}, nil
}

func init() {
	b.Command("/status", func(update *types.Update) error {
		return b.SendMessage(update.Message.Chat.ID, "ok")
	})

	b.Command("/help", func(update *types.Update) error {
		commands, err := b.GetMyCommands()
		if err != nil {
			return err
		}

		var text string
		for _, command := range commands {
			text += fmt.Sprintf("/%s - %s\n", command.Command, command.Description)
		}

		text = strings.TrimRight(text, "\n")

		return b.SendMessage(update.Message.Chat.ID, text)
	})
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
