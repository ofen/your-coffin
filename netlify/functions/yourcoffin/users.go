package main

import (
	"encoding/json"
	"os"

	"github.com/nasermirzaei89/telegram"
)

type user struct {
	ID int `json:"id"`
}

func allowedUsers() []user {
	users := []user{}
	date := []byte(os.Getenv("ALLOWED_USERS"))
	json.Unmarshal(date, &users)

	return users
}

func isAllowed(update *telegram.Update) bool {
	users := allowedUsers()

	for _, user := range users {
		if user.ID == int(update.Message.From.ID) {
			return true
		}
	}

	return false
}
