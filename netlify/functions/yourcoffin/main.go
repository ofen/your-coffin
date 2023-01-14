package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/ofen/yourcoffin/internal/bot"
	"github.com/ofen/yourcoffin/internal/bot/types"
	"github.com/ofen/yourcoffin/internal/db"
	"github.com/ofen/yourcoffin/internal/googlesheets"
)

var (
	b      = bot.New(os.Getenv("BOT_TOKEN"))
	gs     = googlesheets.New(os.Getenv("GOOGLE_SPREADSHEET"))
	secret = os.Getenv("SECRET_TOKEN")
	redis  = db.New(os.Getenv("REDIS_ADDR"), os.Getenv("REDIS_PASSWORD"), time.Minute*5)
)

func main() {
	b.Command("/status", statusHandler)
	b.Command("/help", helpHandler)
	b.Command("/lastmeters", lastmetersHandler)
	b.Command("/meters", metersHandler)

	lambda.Start(handler)
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	log.Println(event)

	lc, ok := lambdacontext.FromContext(ctx)
	if !ok {
		log.Println(lc)
	}

	if header := event.Headers[bot.HeaderSecretToken]; header != secret {
		return &events.APIGatewayProxyResponse{StatusCode: http.StatusMethodNotAllowed}, nil
	}

	if event.HTTPMethod != http.MethodPost {
		return &events.APIGatewayProxyResponse{StatusCode: http.StatusMethodNotAllowed}, nil
	}

	update := &types.Update{}
	if err := json.Unmarshal([]byte(event.Body), update); err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusUnprocessableEntity,
			Body:       err.Error(),
		}, nil
	}

	if err := b.HandleUpdate(ctx, update); err != nil {
		log.Println(err)
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, nil
	}

	return &events.APIGatewayProxyResponse{StatusCode: http.StatusOK}, nil
}
