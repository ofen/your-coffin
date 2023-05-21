package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/nasermirzaei89/telegram"
)

const (
	metersDateFmt string = "02.01.2006"
	// headerSecretToken is secret token header configured via https://core.telegram.org/bots/api#setwebhook
	headerSecretToken = "x-telegram-bot-api-secret-token"

	parseModeHTML       = "HTML"
	parseModeMarkdown   = "Markdown"
	parseModeMarkdownV2 = "MarkdownV2"
)

// handler is main entrypoint for request.
func handler(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	log.Println(event)

	lc, ok := lambdacontext.FromContext(ctx)
	if ok {
		log.Println(lc)
	}

	if header := event.Headers[headerSecretToken]; header != secretToken && event.HTTPMethod != http.MethodPost {
		return &events.APIGatewayProxyResponse{StatusCode: http.StatusNotFound}, nil
	}

	update := &telegram.Update{}
	if err := json.Unmarshal([]byte(event.Body), update); err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusUnprocessableEntity,
			Body:       err.Error(),
		}, nil
	}

	var err error
	switch *update.Message.Text {
	case "/status":
		err = statusHandler(ctx, update)
	case "/help":
		err = helpHandler(ctx, update)
	case "/lastmeters":
		err = lastmetersHandler(ctx, update)
	case "/meters":
		err = metersHandler(ctx, update)
	}

	if err != nil {
		_ = sendMessage(ctx, update.Message.From.ID, err.Error())
	}

	return &events.APIGatewayProxyResponse{StatusCode: http.StatusOK}, nil
}

func _sendMessage(ctx context.Context, opts ...telegram.MethodOption) error {
	resp, err := bot.SendMessage(ctx, opts...)
	if err != nil {
		return err
	}

	if !resp.IsOK() {
		return fmt.Errorf("%d: %s", resp.GetErrorCode(), resp.GetDescription())
	}

	return nil
}

func sendMessage(ctx context.Context, chatID int64, text string) error {
	opts := []telegram.MethodOption{
		telegram.SetChatID(chatID),
		telegram.SetText(text),
	}

	return _sendMessage(ctx, opts...)
}

func sendMessageMarkdownV2(ctx context.Context, chatID int64, text string) error {
	opts := []telegram.MethodOption{
		telegram.SetChatID(chatID),
		telegram.SetText(escapeText(parseModeMarkdownV2, text)),
		telegram.SetParseMode(parseModeMarkdownV2),
	}

	return _sendMessage(ctx, opts...)
}

func statusHandler(ctx context.Context, update *telegram.Update) error {
	return sendMessage(ctx, update.Message.From.ID, "ok")
}

func helpHandler(ctx context.Context, update *telegram.Update) error {
	resp, err := bot.GetMyCommands(ctx)
	if err != nil {
		return err
	}

	var text string
	for _, cmd := range resp.GetCommands() {
		text += fmt.Sprintf("/%s - %s\n", cmd.Command, cmd.Description)
	}

	text = strings.TrimRight(text, "\n")

	return sendMessage(ctx, update.Message.From.ID, text)
}

func lastmetersHandler(ctx context.Context, update *telegram.Update) error {
	if !isAllowed(update) {
		return nil
	}

	v, err := gs.Rows()
	if err != nil {
		return err
	}

	m1 := Rtom(v.Values[len(v.Values)-1])
	text := fmt.Sprintf("**here is the last meters**\n"+
		"date: %v\n"+
		"hot water: %v\n"+
		"cold water: %v\n"+
		"electricity (t1): %v\n"+
		"electricity (t2): %v",
		m1.Date,
		m1.HotWater,
		m1.ColdWater,
		m1.ElectricityT1,
		m1.ElectricityT2,
	)

	if len(v.Values) > 1 {
		m2 := Rtom(v.Values[len(v.Values)-2])
		subm := m1.Sub(m2)

		text = fmt.Sprintf("**here is the last meters**\n"+
			"date: %s\n"+
			"hot water: %d (%+d)\n"+
			"cold water: %d (%+d)\n"+
			"electricity (t1): %d (%+d)\n"+
			"electricity (t2): %d (%+d)",
			m1.Date,
			m1.HotWater, subm.HotWater,
			m1.ColdWater, subm.ColdWater,
			m1.ElectricityT1, subm.ElectricityT1,
			m1.ElectricityT2, subm.ElectricityT2,
		)
	}

	return sendMessageMarkdownV2(ctx, update.Message.From.ID, text)
}

func metersHandlerV2(ctx context.Context, update *telegram.Update) error {
	if !isAllowed(update) {
		return nil
	}

	args := strings.Fields(*update.Message.Text)
	if len(args) < 2 {
		return fmt.Errorf("usage: /meters <hot_water>,<cold_water>,<electricity_t1>,<electricity_t2>")
	}

	values := strings.Split(args[1], ",")
	if len(values) != 4 {
		return fmt.Errorf("invalid argument")
	}

	for _, v := range values {
		if _, err := strconv.Atoi(v); err != nil {
			return err
		}
	}

	lastRows, err := gs.LastRow()
	if err != nil {
		return err
	}

	previousMeters := Rtom(lastRows)
	newMeters := Rtom([]interface{}{time.Unix(int64(update.Message.Date), 0).Format(metersDateFmt), values[0], values[1], values[2], values[3]})

	err = gs.AppendRow(Mtor(newMeters))
	if err != nil {
		return err
	}

	subMeters := newMeters.Sub(previousMeters)

	text := fmt.Sprintf("**meters updated**\n"+
		"date: %s\n"+
		"hot water: %d (%+d)\n"+
		"cold water: %d (%+d)\n"+
		"electricity (t1): %d (%+d)\n"+
		"electricity (t2): %d (%+d)",
		newMeters.Date,
		newMeters.HotWater, subMeters.HotWater,
		newMeters.ColdWater, subMeters.ColdWater,
		newMeters.ElectricityT1, subMeters.ElectricityT1,
		newMeters.ElectricityT2, subMeters.ElectricityT2,
	)

	return sendMessageMarkdownV2(ctx, update.Message.From.ID, text)
}

func metersHandler(ctx context.Context, update *telegram.Update) error {
	if !isAllowed(update) {
		return nil
	}

	args := strings.Fields(*update.Message.Text)
	if len(args) < 2 {
		return fmt.Errorf("usage: /meters <hot_water>,<cold_water>,<electricity_t1>,<electricity_t2>")
	}

	values := strings.Split(args[1], ",")
	if len(values) != 4 {
		return fmt.Errorf("invalid argument")
	}

	for _, v := range values {
		if _, err := strconv.Atoi(v); err != nil {
			return err
		}
	}

	lastRows, err := gs.LastRow()
	if err != nil {
		return err
	}

	previousMeters := Rtom(lastRows)
	newMeters := Rtom([]interface{}{time.Unix(int64(update.Message.Date), 0).Format(metersDateFmt), values[0], values[1], values[2], values[3]})

	err = gs.AppendRow(Mtor(newMeters))
	if err != nil {
		return err
	}

	subMeters := newMeters.Sub(previousMeters)

	text := fmt.Sprintf("*meters updated*\n"+
		"date: %s\n"+
		"hot water: %d (%+d)\n"+
		"cold water: %d (%+d)\n"+
		"electricity (t1): %d (%+d)\n"+
		"electricity (t2): %d (%+d)",
		newMeters.Date,
		newMeters.HotWater, subMeters.HotWater,
		newMeters.ColdWater, subMeters.ColdWater,
		newMeters.ElectricityT1, subMeters.ElectricityT1,
		newMeters.ElectricityT2, subMeters.ElectricityT2,
	)

	return sendMessageMarkdownV2(ctx, update.Message.From.ID, text)
}

func escapeText(parseMode string, text string) string {
	var replacer *strings.Replacer

	switch parseMode {
	case parseModeHTML:
		replacer = strings.NewReplacer("<", "&lt;", ">", "&gt;", "&", "&amp;")
	case parseModeMarkdown:
		replacer = strings.NewReplacer("_", "\\_", "*", "\\*", "`", "\\`", "[", "\\[")
	case parseModeMarkdownV2:
		replacer = strings.NewReplacer(
			"_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]", "(",
			"\\(", ")", "\\)", "~", "\\~", "`", "\\`", ">", "\\>",
			"#", "\\#", "+", "\\+", "-", "\\-", "=", "\\=", "|",
			"\\|", "{", "\\{", "}", "\\}", ".", "\\.", "!", "\\!",
		)
	default:
		return text
	}

	return replacer.Replace(text)
}
