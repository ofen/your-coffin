package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ofen/yourcoffin/internal/bot/types"
)

var (
	statusHandler = func(update *types.Update) error {
		_, err := b.SendMessage(update.Message.Chat.ID, "ok")

		return err
	}

	helpHandler = func(update *types.Update) error {
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

	lastmetersHandler = func(update *types.Update) error {
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
	}

	metersHandler = func(update *types.Update) error {
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
		)

		return err
	}
)
