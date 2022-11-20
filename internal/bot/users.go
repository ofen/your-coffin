package bot

import (
	"encoding/json"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type User struct {
	ID int64 `json:"id"`
}

func GetAllowedUsers() []User {
	users := []User{}
	date := []byte(os.Getenv("ALLOWED_USERS"))
	json.Unmarshal(date, &users)

	return users
}

func IsAllowed(message *tgbotapi.Message) bool {
	users := GetAllowedUsers()

	for _, user := range users {
		if user.ID == message.Chat.ID {
			return true
		}
	}

	return false
}
