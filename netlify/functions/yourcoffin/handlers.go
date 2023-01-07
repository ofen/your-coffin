package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ofen/yourcoffin/internal/bot/types"
)

const metersDateFmt string = "02.01.2006"

var statusHandler = func(update *types.Update) error {
	_, err := b.SendMessage(update.Message.Chat.ID, "ok", types.ParseModeMarkdownV2)

	return err
}

var helpHandler = func(update *types.Update) error {
	commands, err := b.GetMyCommands()
	if err != nil {
		_, err = b.SendMessage(update.Message.Chat.ID, err.Error(), types.ParseModeMarkdownV2)

		return err
	}

	var text string
	for _, command := range commands {
		text += fmt.Sprintf("/%s - %s\n", command.Command, command.Description)
	}

	text = strings.TrimRight(text, "\n")

	_, err = b.SendMessage(update.Message.Chat.ID, text, types.ParseModeMarkdownV2)

	return err
}

var lastmetersHandler = func(update *types.Update) error {
	if !IsAllowed(update) {
		return nil
	}

	v, err := gs.Rows()
	if err != nil {
		_, err = b.SendMessage(update.Message.Chat.ID, err.Error(), types.ParseModeMarkdownV2)

		return err
	}

	m1 := Rtom(v.Values[len(v.Values)-1])
	text := fmt.Sprintf("*here is the last meters*"+
		"\n\ndate: %v"+
		"\nhot water: %v"+
		"\ncold water: %v"+
		"\nelectricity (t1): %v"+
		"\nelectricity (t2): %v",
		m1.Date,
		m1.HotWater,
		m1.ColdWater,
		m1.ElectricityT1,
		m1.ElectricityT2,
	)

	if len(v.Values) > 1 {
		m2 := Rtom(v.Values[len(v.Values)-2])
		subm := m1.Sub(m2)

		text = fmt.Sprintf("*meters updated*"+
			"\n\ndate: %s"+
			"\nhot water: %d (%+d)"+
			"\ncold water: %d (%+d)"+
			"\nelectricity (t1): %d (%+d)"+
			"\nelectricity (t2): %d (%+d)",
			m1.Date,
			m1.HotWater, subm.HotWater,
			m1.ColdWater, subm.ColdWater,
			m1.ElectricityT1, subm.ElectricityT1,
			m1.ElectricityT2, subm.ElectricityT2,
		)
	}

	_, err = b.SendMessage(update.Message.Chat.ID, text, types.ParseModeMarkdownV2)

	return err
}

var metersHandler = func(update *types.Update) error {
	if !IsAllowed(update) {
		return nil
	}

	args := update.Message.Args()
	if len(args) < 2 {
		_, err := b.SendMessage(update.Message.Chat.ID, "usage: /meters <hot_water>,<cold_water>,<electricity_t1>,<electricity_t2>", types.ParseModeMarkdownV2)

		return err
	}

	values := strings.Split(args[1], ",")
	if len(values) != 4 {
		_, err := b.SendMessage(update.Message.Chat.ID, "invalid argument", types.ParseModeMarkdownV2)

		return err
	}

	for _, v := range values {
		_, err := strconv.Atoi(v)
		if err != nil {
			_, err = b.SendMessage(update.Message.Chat.ID, err.Error(), types.ParseModeMarkdownV2)

			return err
		}
	}

	lastRows, err := gs.LastRow()
	if err != nil {
		_, err = b.SendMessage(update.Message.Chat.ID, err.Error(), types.ParseModeMarkdownV2)

		return err
	}

	previousMeters := Rtom(lastRows)

	newMeters := Rtom([]interface{}{update.Message.Date.Format(metersDateFmt), values[0], values[1], values[2], values[3]})

	err = gs.AppendRow(Mtor(newMeters))
	if err != nil {
		_, err = b.SendMessage(update.Message.Chat.ID, err.Error(), types.ParseModeMarkdownV2)

		return err
	}

	_, err = b.SendMessage(
		update.Message.Chat.ID,
		fmt.Sprintf("*meters updated*"+
			"\n\ndate: %s"+
			"\nhot water: %d (%+d)"+
			"\ncold water: %d (%+d)"+
			"\nelectricity (t1): %d (%+d)"+
			"\nelectricity (t2): %d (%+d)",
			newMeters.Date,
			newMeters.HotWater, newMeters.HotWater-previousMeters.HotWater,
			newMeters.ColdWater, newMeters.ColdWater-previousMeters.ColdWater,
			newMeters.ElectricityT1, newMeters.ElectricityT1-previousMeters.ElectricityT1,
			newMeters.ElectricityT2, newMeters.ElectricityT2-previousMeters.ElectricityT2,
		),
		types.ParseModeMarkdownV2,
	)

	return err
}
