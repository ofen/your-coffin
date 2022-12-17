package bot

import (
	"testing"

	"github.com/ofen/yourcoffin/internal/bot/types"
)

func TestHandleUpdate(t *testing.T) {
	args := []*types.Update{
		{
			UpdateID: 1234567890,
			Message: types.Message{
				MessageID: 1,
				From: types.User{
					ID:           1234567890,
					IsBot:        false,
					FirstName:    "User",
					Username:     "user",
					LanguageCode: "en",
				},
				Chat: types.Chat{
					ID:        1234567890,
					FirstName: "User",
					Username:  "user",
					Type:      "private",
				},
				Date: 1669545311,
				Text: "/test",
				Entities: []types.MessageEntity{
					{
						Type:    types.MessageEntityBotCommand,
						Offeset: 0,
						Length:  5,
					},
				},
			},
		},
		{
			UpdateID: 1234567890,
			Message: types.Message{
				MessageID: 1,
				From: types.User{
					ID:           1234567890,
					IsBot:        false,
					FirstName:    "User",
					Username:     "user",
					LanguageCode: "en",
				},
				Chat: types.Chat{
					ID:        1234567890,
					FirstName: "User",
					Username:  "user",
					Type:      "private",
				},
				Date: 1669545311,
				Text: "/test first second third",
				Entities: []types.MessageEntity{
					{
						Type:    types.MessageEntityBotCommand,
						Offeset: 0,
						Length:  5,
					},
				},
			},
		},
		{
			UpdateID: 1234567890,
			Message: types.Message{
				MessageID: 1,
				From: types.User{
					ID:           1234567890,
					IsBot:        false,
					FirstName:    "User",
					Username:     "user",
					LanguageCode: "en",
				},
				Chat: types.Chat{
					ID:        1234567890,
					FirstName: "User",
					Username:  "user",
					Type:      "private",
				},
				Date: 1669545311,
				Text: "test first second third",
			},
		},
	}

	b := New("")
	b.Command("/test", func(update *types.Update) error {
		t.Log(update.Message.Args())

		return nil
	})

	for _, u := range args {
		if err := b.HandleCommand(u); err != nil {
			t.Fatal(err)
		}
	}
}
