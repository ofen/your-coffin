package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

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

func init() {
	b.Command("/status", statusHandler)
	b.Command("/help", helpHandler)
	b.Command("/lastmeters", lastmetersHandler)
	b.Command("/meters", metersHandler)
}

func main() {
	lambda.Start(handler)
}

type Meters struct {
	Date          string
	HotWater      int
	ColdWater     int
	ElectricityT1 int
	ElectricityT2 int
}

func Rtom(row []interface{}) *Meters {
	m := &Meters{}

	m.Date = row[0].(string)
	m.HotWater, _ = strconv.Atoi(row[1].(string))
	m.ColdWater, _ = strconv.Atoi(row[2].(string))
	m.ElectricityT1, _ = strconv.Atoi(row[3].(string))
	m.ElectricityT2, _ = strconv.Atoi(row[4].(string))

	return m
}

func Mtor(m *Meters) []interface{} {
	return []interface{}{m.Date, m.HotWater, m.ColdWater, m.ElectricityT1, m.ElectricityT2}
}

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
