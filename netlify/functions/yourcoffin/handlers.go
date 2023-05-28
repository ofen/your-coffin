package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
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

type Session struct {
	Hanlder string
	Data    json.RawMessage
}

// handler is main entrypoint for request.
func handler(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("recovered: %s\n%s", r, debug.Stack())
		}
	}()

	log.Println(event)

	lc, ok := lambdacontext.FromContext(ctx)
	if ok {
		log.Println(lc)
	}

	if header := event.Headers[headerSecretToken]; header != secretToken && event.HTTPMethod != http.MethodPost {
		return &events.APIGatewayProxyResponse{StatusCode: http.StatusNotFound}, nil
	}

	u := &update{}
	if err := json.Unmarshal([]byte(event.Body), u); err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusUnprocessableEntity,
			Body:       err.Error(),
		}, nil
	}

	if !isAllowed(u) {
		return &events.APIGatewayProxyResponse{StatusCode: http.StatusOK}, nil
	}

	var session Session
	if err := s.Get(ctx, strconv.FormatInt(u.Message.From.ID, 10), &session); err != nil {
		log.Println(err)
	}

	if session.Hanlder == "" {
		session.Hanlder = u.command()
	}

	ctx = context.WithValue(ctx, "session", session)

	var err error
	switch session.Hanlder {
	case "/status":
		err = statusHandler(ctx, u)
	case "/help":
		err = helpHandler(ctx, u)
	case "/lastmeters":
		err = lastmetersHandler(ctx, u)
	case "/meters":
		err = metersHandler(ctx, u)
	default:
		err = sendMessage(ctx, u.Message.From.ID, "incorrect command")
	}

	if err != nil {
		if err = sendMessage(ctx, u.Message.From.ID, err.Error()); err != nil {
			log.Println(err)
		}
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

func statusHandler(ctx context.Context, u *update) error {
	return sendMessage(ctx, u.Message.From.ID, "ok")
}

func helpHandler(ctx context.Context, u *update) error {
	resp, err := bot.GetMyCommands(ctx)
	if err != nil {
		return err
	}

	var text string
	for _, cmd := range resp.GetCommands() {
		text += fmt.Sprintf("/%s - %s\n", cmd.Command, cmd.Description)
	}

	text = strings.TrimRight(text, "\n")

	return sendMessage(ctx, u.Message.From.ID, text)
}

func lastmetersHandler(ctx context.Context, u *update) error {
	l, err := listMeters()
	if err != nil {
		return err
	}

	m1 := l[len(l)-1]
	text := fmt.Sprintf("*here is the last meters*\n"+
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

	if len(l) > 2 {
		m2 := l[len(l)-2]
		subm := m1.sub(m2)

		text = fmt.Sprintf("*here is the last meters*\n"+
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

	return sendMessageMarkdownV2(ctx, u.Message.From.ID, text)
}

func metersHandler(ctx context.Context, u *update) error {
	session, _ := ctx.Value("session").(Session)
	if session.Data == nil {
		data, err := json.Marshal(&meters{})
		if err != nil {
			return err
		}

		session.Data = data

		if err := sendMessage(ctx, u.Message.From.ID, "enter hot water or \"cancel\" to cancel"); err != nil {
			return err
		}

		return s.Set(ctx, strconv.FormatInt(u.Message.From.ID, 10), session)
	}

	if *u.Message.Text == "cancel" {
		if err := sendMessage(ctx, u.Message.From.ID, "aborted"); err != nil {
			return err
		}

		return s.Del(ctx, strconv.FormatInt(u.Message.From.ID, 10))
	}

	m := &meters{}
	if err := json.Unmarshal(session.Data, m); err != nil {
		return err
	}

	if m.HotWater == 0 {
		v, err := strconv.Atoi(*u.Message.Text)
		if err != nil {
			return fmt.Errorf("invalid argument: %q", *u.Message.Text)
		}

		m.HotWater = v

		data, err := json.Marshal(m)
		if err != nil {
			return err
		}

		session.Data = data

		if err := sendMessage(ctx, u.Message.From.ID, "enter cold water"); err != nil {
			return err
		}

		return s.Set(ctx, strconv.FormatInt(u.Message.From.ID, 10), session)
	}

	if m.ColdWater == 0 {
		v, err := strconv.Atoi(*u.Message.Text)
		if err != nil {
			return fmt.Errorf("invalid argument: %q", *u.Message.Text)
		}

		m.ColdWater = v

		data, err := json.Marshal(m)
		if err != nil {
			return err
		}

		session.Data = data

		if err := sendMessage(ctx, u.Message.From.ID, "enter electricity (t1)"); err != nil {
			return err
		}

		return s.Set(ctx, strconv.FormatInt(u.Message.From.ID, 10), session)
	}

	if m.ElectricityT1 == 0 {
		v, err := strconv.Atoi(*u.Message.Text)
		if err != nil {
			return fmt.Errorf("invalid argument: %q", *u.Message.Text)
		}

		m.ElectricityT1 = v

		data, err := json.Marshal(m)
		if err != nil {
			return err
		}

		session.Data = data

		if err := sendMessage(ctx, u.Message.From.ID, "enter electricity (t2)"); err != nil {
			return err
		}

		return s.Set(ctx, strconv.FormatInt(u.Message.From.ID, 10), session)
	}

	if m.ElectricityT2 == 0 {
		v, err := strconv.Atoi(*u.Message.Text)
		if err != nil {
			return fmt.Errorf("invalid argument: %q", *u.Message.Text)
		}

		m.ElectricityT2 = v

		data, err := json.Marshal(m)
		if err != nil {
			return err
		}

		session.Data = data

		prevm, err := lastMeters()
		if err != nil {
			return err
		}

		subm := m.sub(prevm)

		if err = sendMessageMarkdownV2(ctx, u.Message.From.ID, fmt.Sprintf("is this correct?\n"+
			"hot water: %d (%+d)\n"+
			"cold water: %d (%+d)\n"+
			"electricity (t1): %d (%+d)\n"+
			"electricity (t2): %d (%+d)",
			m.HotWater, subm.HotWater,
			m.ColdWater, subm.ColdWater,
			m.ElectricityT1, subm.ElectricityT1,
			m.ElectricityT2, subm.ElectricityT2,
		)); err != nil {
			return err
		}

		return s.Set(ctx, strconv.FormatInt(u.Message.From.ID, 10), session)
	}

	if *u.Message.Text != "ok" {
		return fmt.Errorf("only \"ok\" or \"cancel\" allowed")
	}

	m.Date = u.Time().Format(metersDateFmt)

	if err := appendMeters(m); err != nil {
		return err
	}

	if err := s.Del(ctx, strconv.FormatInt(u.Message.From.ID, 10)); err != nil {
		return err
	}

	return sendMessage(ctx, u.Message.From.ID, "meters updated")

}

func escapeText(parseMode string, text string) string {
	var replacer *strings.Replacer

	switch parseMode {
	case parseModeHTML:
		replacer = strings.NewReplacer(
			"<", "&lt;",
			">", "&gt;",
			"&", "&amp;",
		)
	case parseModeMarkdown:
		replacer = strings.NewReplacer(
			"_", "\\_",
			"*", "\\*",
			"`", "\\`",
			"[", "\\[",
		)
	case parseModeMarkdownV2:
		replacer = strings.NewReplacer(
			"_", "\\_",
			"[", "\\[",
			"]", "\\]",
			"(", "\\(",
			")", "\\)",
			"~", "\\~",
			"`", "\\`",
			">", "\\>",
			"#", "\\#",
			"+", "\\+",
			"-", "\\-",
			"=", "\\=",
			"|", "\\|",
			"{", "\\{",
			"}", "\\}",
			".", "\\.",
			"!", "\\!",
		)
	default:
		return text
	}

	return replacer.Replace(text)
}

type update struct {
	telegram.Update
}

func (u update) command() string {
	args := u.args()
	if len(args) > 0 {
		return args[0]
	}

	return ""
}

func (u update) args() []string {
	return strings.Fields(*u.Message.Text)
}

func (u update) Time() time.Time {
	return time.Unix(int64(u.Message.Date), 0)
}
