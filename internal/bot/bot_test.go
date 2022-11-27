package bot

import (
	"testing"
)

func TestHandleUpdate(t *testing.T) {
	args := []*Update{
		{
			UpdateID: 1234567890,
			Message: Message{
				MessageID: 1,
				From: User{
					ID:           1234567890,
					IsBot:        false,
					FirstName:    "User",
					Username:     "user",
					LanguageCode: "en",
				},
				Chat: Chat{
					ID:        1234567890,
					FirstName: "User",
					Username:  "user",
					Type:      "private",
				},
				Date: 1669545311,
				Text: "/test",
				Entities: []MessageEntity{
					{
						Type:    MessageEntityBotCommand,
						Offeset: 0,
						Length:  5,
					},
				},
			},
		},
		{
			UpdateID: 1234567890,
			Message: Message{
				MessageID: 1,
				From: User{
					ID:           1234567890,
					IsBot:        false,
					FirstName:    "User",
					Username:     "user",
					LanguageCode: "en",
				},
				Chat: Chat{
					ID:        1234567890,
					FirstName: "User",
					Username:  "user",
					Type:      "private",
				},
				Date: 1669545311,
				Text: "/test first second third",
				Entities: []MessageEntity{
					{
						Type:    MessageEntityBotCommand,
						Offeset: 0,
						Length:  5,
					},
				},
			},
		},
		{
			UpdateID: 1234567890,
			Message: Message{
				MessageID: 1,
				From: User{
					ID:           1234567890,
					IsBot:        false,
					FirstName:    "User",
					Username:     "user",
					LanguageCode: "en",
				},
				Chat: Chat{
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
	b.Command("/test", func(update *Update) error {
		t.Log(update.Message.Args())

		return nil
	})

	for _, u := range args {
		if err := b.HandleCommand(u); err != nil {
			t.Fatal(err)
		}
	}
}
