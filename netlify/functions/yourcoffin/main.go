package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/ofen/yourcoffin/internal/bot"
	"github.com/ofen/yourcoffin/internal/bot/types"

	"github.com/ofen/yourcoffin/internal/googlesheets"
)

var (
	b      = bot.New(os.Getenv("BOT_TOKEN"))
	gs     = googlesheets.New(os.Getenv("GOOGLE_SPREADSHEET"))
	secret = os.Getenv("SECRET_TOKEN")
)

func main() {
	b.Command("/status", statusHandler)
	b.Command("/help", helpHandler)
	b.Command("/lastmeters", lastmetersHandler)
	b.Command("/meters", metersHandler)

	lambda.Start(handler)
}

func handler(r events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	log.Println(r)

	if header := r.Headers[bot.HeaderSecretToken]; header != secret {
		return &events.APIGatewayProxyResponse{StatusCode: http.StatusMethodNotAllowed}, nil
	}

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

	ctx := context.Background()
	if err := b.HandleUpdate(ctx, update); err != nil {
		log.Println(err)
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, nil
	}

	return &events.APIGatewayProxyResponse{StatusCode: http.StatusOK}, nil
}
