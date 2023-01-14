package main

import (
	"encoding/json"
	"os"

	"github.com/ofen/yourcoffin/internal/bot/types"
)

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
