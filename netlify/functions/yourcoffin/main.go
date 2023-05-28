package main

import (
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/nasermirzaei89/telegram"
	"github.com/ofen/yourcoffin/internal/googlesheets"
	"github.com/ofen/yourcoffin/internal/storage"
)

var (
	bot         = telegram.New(os.Getenv("BOT_TOKEN"))
	gs          = googlesheets.New(os.Getenv("GOOGLE_SPREADSHEET"))
	secretToken = os.Getenv("SECRET_TOKEN")
	s           = storage.New(os.Getenv("REDIS_ADDR"), os.Getenv("REDIS_PASSWORD"), time.Minute*5)
)

func main() {
	lambda.Start(handler)
}
