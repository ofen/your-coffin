package bot

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const keyMeters contextKey = "meters"

func withMeters(ctx context.Context, m meters) context.Context {
	return context.WithValue(ctx, keyMeters, m)
}

func getMeters(ctx context.Context) (meters, bool) {
	v, ok := ctx.Value(keyMeters).(meters)
	return v, ok
}

func (b *Bot) handleError(message *tg.Message, err error) {
	log.Println(err)
	msg := tg.NewMessage(message.Chat.ID, err.Error())
	b.Send(msg)
}

func (b *Bot) updateMeters(message *tg.Message, m meters) {
	s, err := newSpreadsheet(os.Getenv("GOOGLE_SPREADSHEET"))
	if err != nil {
		b.handleError(message, err)
		return
	}

	data, err := s.lastRow()
	if err != nil {
		b.handleError(message, err)
	}

	hotWater, _ := strconv.Atoi(data[1].(string))
	coldWater, _ := strconv.Atoi(data[2].(string))
	electricityT1, _ := strconv.Atoi(data[3].(string))
	electricityT2, _ := strconv.Atoi(data[4].(string))

	previousMeters := meters{
		hotWater:      hotWater,
		coldWater:     coldWater,
		electricityT1: electricityT1,
		electricityT2: electricityT2,
	}

	today := message.Time().Format(metersDateFormat)
	err = s.appendRow([]interface{}{
		today,
		m.hotWater,
		m.coldWater,
		m.electricityT1,
		m.electricityT2,
	})
	if err != nil {
		b.Send(tg.NewMessage(message.Chat.ID, err.Error()))
		return
	}
	msg := tg.NewMessage(
		message.Chat.ID,
		fmt.Sprintf(
			"sheet updated"+
				"\n\n*hot water:* %d (%+d)"+
				"\n*cold water:* %d (%+d)"+
				"\n*electricity (t1):* %d (%+d)"+
				"\n*electricity (t2):* %d (%+d)",
			m.hotWater, m.hotWater-previousMeters.hotWater,
			m.coldWater, m.coldWater-previousMeters.coldWater,
			m.electricityT1, m.electricityT1-previousMeters.electricityT1,
			m.electricityT2, m.electricityT2-previousMeters.electricityT2,
		),
	)
	msg.ParseMode = tg.ModeMarkdown
	b.Send(msg)
}

func (b *Bot) WhoAmIHandler(message *tg.Message) {
	text := fmt.Sprintf("*id:* %v\n*name:* %s %s", message.Chat.ID, message.Chat.FirstName, message.Chat.LastName)
	msg := tg.NewMessage(message.Chat.ID, text)
	msg.ParseMode = tg.ModeMarkdown
	b.Send(msg)
}
func (b *Bot) StatusHandler(message *tg.Message) {
	text := ""
	version := os.Getenv("SOURCE_VERSION")

	switch {
	case version != "":
		text = fmt.Sprintf("version: %s", version)
	default:
		text = "ok"
	}

	msg := tg.NewMessage(message.Chat.ID, text)
	b.Send(msg)
}
func (b *Bot) MetersHandler(message *tg.Message) {
	b.Send(tg.NewMessage(message.Chat.ID, "enter hot water or use /cancel to stop"))
	b.RegisterNextStepHandler(message, b.handleHotWater(meters{-1, -1, -1, -1}))
}

func (b *Bot) handleHotWater(m meters) handlerFunc {
	return func(message *tg.Message) {
		if message.IsCommand() && message.Command() == "cancel" {
			b.Send(tg.NewMessage(message.Chat.ID, "canceled"))
			return
		}
		v, err := strconv.Atoi(message.Text)
		if err != nil || v < 0 {
			b.handleError(message, fmt.Errorf("value should be positive number"))
			b.RegisterNextStepHandler(message, b.handleHotWater(m))
			return
		}

		m.hotWater = v
		b.Send(tg.NewMessage(message.Chat.ID, "enter cold water"))
		b.RegisterNextStepHandler(message, b.handleColdWater(m))
	}
}

func (b *Bot) handleColdWater(m meters) handlerFunc {
	return func(message *tg.Message) {
		if message.IsCommand() && message.Command() == "cancel" {
			b.Send(tg.NewMessage(message.Chat.ID, "canceled"))
			return
		}
		v, err := strconv.Atoi(message.Text)
		if err != nil || v < 0 {
			b.handleError(message, fmt.Errorf("value should be positive number"))
			b.RegisterNextStepHandler(message, b.handleColdWater(m))
			return
		}

		m.coldWater = v
		b.Send(tg.NewMessage(message.Chat.ID, "enter electricity (t1)"))
		b.RegisterNextStepHandler(message, b.handleElectricityT1(m))
	}
}

func (b *Bot) handleElectricityT1(m meters) handlerFunc {
	return func(message *tg.Message) {
		if message.IsCommand() && message.Command() == "cancel" {
			b.Send(tg.NewMessage(message.Chat.ID, "canceled"))
			return
		}
		v, err := strconv.Atoi(message.Text)
		if err != nil || v < 0 {
			b.handleError(message, fmt.Errorf("value should be positive number"))
			b.RegisterNextStepHandler(message, b.handleElectricityT1(m))
			return
		}

		m.electricityT1 = v
		b.Send(tg.NewMessage(message.Chat.ID, "enter electricity (t2)"))
		b.RegisterNextStepHandler(message, b.handleElectricityT2(m))
	}
}

func (b *Bot) handleElectricityT2(m meters) handlerFunc {
	return func(message *tg.Message) {
		if message.IsCommand() && message.Command() == "cancel" {
			b.Send(tg.NewMessage(message.Chat.ID, "canceled"))
			return
		}
		v, err := strconv.Atoi(message.Text)
		if err != nil || v < 0 {
			b.handleError(message, fmt.Errorf("value should be positive number"))
			b.RegisterNextStepHandler(message, b.handleElectricityT2(m))
			return
		}

		m.electricityT2 = v

		b.updateMeters(message, m)
	}
}

func (b *Bot) LastmetersHandler(message *tg.Message) {
	s, err := newSpreadsheet(os.Getenv("GOOGLE_SPREADSHEET"))
	if err != nil {
		b.Send(tg.NewMessage(message.Chat.ID, err.Error()))
		return
	}

	v, err := s.lastRow()
	if err != nil {
		b.Send(tg.NewMessage(message.Chat.ID, err.Error()))
		return
	}

	msg := tg.NewMessage(
		message.Chat.ID,
		fmt.Sprintf(
			"here is the last meters"+
				"\n\n*date:* %v"+
				"\n*hot water:* %v"+
				"\n*cold water:* %v"+
				"\n*electricity (t1):* %v"+
				"\n*electricity (t2):* %v",
			v[:5]...,
		),
	)
	msg.ParseMode = tg.ModeMarkdown
	b.Send(msg)
}

func (b *Bot) UnsupportedHandler(message *tg.Message) {
	b.handleError(message, fmt.Errorf("unsupported command: %q", message.Command()))
}

func (b *Bot) CurrencyHandler(message *tg.Message) {
	resp, err := http.Get(curencyEndpoint)
	if err != nil {
		b.handleError(message, err)
		return
	}
	defer resp.Body.Close()

	v := &valCurs{}
	if err = newXMLDecoder(resp.Body).Decode(&v); err != nil {
		b.handleError(message, err)
		return
	}

	report := []string{}

	for _, valute := range v.Valute {
		switch strings.ToLower(valute.CharCode) {
		case "usd", "eur":
			valuteValue, err := strconv.ParseFloat(strings.Replace(valute.Value, ",", ".", 1), 64)
			if err != nil {
				log.Fatalln(err)
			}
			report = append(report, fmt.Sprintf("*%s:* %.2f", valute.CharCode, valuteValue))
		}

	}

	text := strings.Join(report, "\n")

	if text == "" {
		text = "No exchange rate found"
	}

	msg := tg.NewMessage(message.Chat.ID, text)
	msg.ParseMode = tg.ModeMarkdown
	b.Send(msg)
}
