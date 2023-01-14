package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/ofen/yourcoffin/internal/bot"
	"github.com/ofen/yourcoffin/internal/bot/types"
)

const metersDateFmt string = "02.01.2006"

func statusHandler(ctx context.Context, update *types.Update) error {
	_, err := b.SendMessage(update.Message.Chat.ID, "ok")

	return err
}

func helpHandler(ctx context.Context, update *types.Update) error {
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
}

func lastmetersHandler(ctx context.Context, update *types.Update) error {
	if !IsAllowed(update) {
		return nil
	}

	v, err := gs.Rows()
	if err != nil {
		_, err = b.SendMessage(update.Message.Chat.ID, err.Error())

		return err
	}

	m1 := Rtom(v.Values[len(v.Values)-1])
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

	if len(v.Values) > 1 {
		m2 := Rtom(v.Values[len(v.Values)-2])
		subm := m1.Sub(m2)

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

	_, err = b.SendMessage(update.Message.Chat.ID, text)

	return err
}

func testHandlerThird(ctx context.Context, update *types.Update) error {
	v := ctx.Value("test")
	previousMessage, ok := v.(string)
	if !ok {
		return nil
	}

	b.SendMessage(update.Message.Chat.ID, fmt.Sprintf("you entered: %s and %s", previousMessage, update.Message.Text))

	return nil
}

func testHandlerSecond(ctx context.Context, update *types.Update) error {
	b.SendMessage(update.Message.Chat.ID, "enter something else")

	b.SetNextHandler(update, func(ctx context.Context, update *types.Update) error {
		ctx = context.WithValue(ctx, "test", update.Message.Text)

		return testHandlerThird(ctx, update)
	})

	return nil
}

func testHandler(ctx context.Context, update *types.Update) error {
	b.SendMessage(update.Message.Chat.ID, "enter something")

	b.SetNextHandler(update, testHandlerSecond)

	return nil
}

func metersHandler(ctx context.Context, update *types.Update) error {
	if !IsAllowed(update) {
		return nil
	}

	args := update.Message.Args()
	if len(args) < 2 {
		_, err := b.SendMessage(update.Message.Chat.ID, "usage: /meters <hot_water>,<cold_water>,<electricity_t1>,<electricity_t2>")

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

	lastRows, err := gs.LastRow()
	if err != nil {
		_, err = b.SendMessage(update.Message.Chat.ID, err.Error())

		return err
	}

	previousMeters := Rtom(lastRows)
	newMeters := Rtom([]interface{}{update.Message.Date.Format(metersDateFmt), values[0], values[1], values[2], values[3]})

	err = gs.AppendRow(Mtor(newMeters))
	if err != nil {
		_, err = b.SendMessage(update.Message.Chat.ID, err.Error())

		return err
	}

	subMeters := newMeters.Sub(previousMeters)

	_, err = b.SendMessage(
		update.Message.Chat.ID,
		fmt.Sprintf("*meters updated*\n"+
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
		),
	)

	return err
}
