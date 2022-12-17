package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/ofen/yourcoffin/internal/bot"
	"github.com/ofen/yourcoffin/internal/bot/types"
	"github.com/ofen/yourcoffin/internal/googlesheets"
)

const metersDateFmt string = "02.01.2006"

var (
	b  = bot.New(os.Getenv("BOT_TOKEN"))
	gs = googlesheets.New(os.Getenv("GOOGLE_SPREADSHEET"))
)

type User struct {
	ID int `json:"id"`
}

func AllowedUsers() []User {
	users := []User{}
	date := []byte(os.Getenv("ALLOWED_USERS"))
	json.Unmarshal(date, &users)

	return users
}

func IsAllowed(update *types.Update) bool {
	users := AllowedUsers()

	for _, user := range users {
		if user.ID == update.Message.Chat.ID {
			return true
		}
	}

	return false
}

func handler(r events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	log.Println(r)

	if r.HTTPMethod != http.MethodPost {
		return &events.APIGatewayProxyResponse{StatusCode: http.StatusMethodNotAllowed}, nil
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
		_, err := b.SendMessage(update.Message.Chat.ID, "ok")

		return err
	})

	b.Command("/help", func(update *types.Update) error {
		commands, err := b.GetMyCommands()
		if err != nil {
			_, err = b.SendMessage(update.Message.Chat.ID, err.Error())

			return err
		}

		var text string
		for _, command := range commands {
			text += fmt.Sprintf("/%s - %s\n", command.Command, command.Description)
		}

		text = strings.TrimRight(text, "\n")

		_, err = b.SendMessage(update.Message.Chat.ID, text)

		return err
	})

	b.Command("/lastmeters", func(update *types.Update) error {
		if !IsAllowed(update) {
			return nil
		}

		v, err := gs.LastRow()
		if err != nil {
			_, err = b.SendMessage(update.Message.Chat.ID, err.Error())

			return err
		}

		_, err = b.SendMessage(
			update.Message.Chat.ID,
			fmt.Sprintf(
				"*here is the last meters*"+
					"\n\ndate: %v"+
					"\nhot water: %v"+
					"\ncold water: %v"+
					"\nelectricity (t1): %v"+
					"\nelectricity (t2): %v",
				v[:5]...,
			),
		)

		return err
	})

	b.Command("/meters", func(update *types.Update) error {
		if !IsAllowed(update) {
			return nil
		}

		args := update.Message.Args()
		if len(args) < 2 {
			_, err := b.SendMessage(update.Message.Chat.ID, "cannot be used without arguments")

			return err
		}

		values := strings.Split(args[1], ",")
		if len(values) != 4 {
			_, err := b.SendMessage(update.Message.Chat.ID, "invalid argument")

			return err
		}

		for _, v := range values {
			_, err := strconv.Atoi(v)
			if err != nil {
				_, err = b.SendMessage(update.Message.Chat.ID, err.Error())

				return err
			}
		}

		date := update.Message.Date.Format(metersDateFmt)

		err := gs.AppendRow(date, values)
		if err != nil {
			_, err = b.SendMessage(update.Message.Chat.ID, err.Error())

			return err
		}

		_, err = b.SendMessage(
			update.Message.Chat.ID,
			fmt.Sprintf("*meters updated*"+
				"\n\ndate: %v"+
				"\nhot water: %v"+
				"\ncold water: %v"+
				"\nelectricity (t1): %v"+
				"\nelectricity (t2): %v",
				date, values[0], values[1], values[2], values[3],
			),
		)

		return err
	})
}

func main() {
	lambda.Start(handler)
}
