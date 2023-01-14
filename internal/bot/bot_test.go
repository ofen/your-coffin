package bot

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ofen/yourcoffin/internal/bot/types"
)

func TestHandleUpdate(t *testing.T) {
	args := []string{
		`{"update_id":1234567890,"message":{"message_id":1,"from":{"id":1234567890,"is_bot":false,"first_name":"User","username":"user","language_code":"en"},"chat":{"id":1234567890,"first_name":"User","username":"user","type":"private"},"date":1669545311,"text":"/test","entities":[{"offset":0,"length":5,"type":"bot_command","url":"","user":{"id":0,"is_bot":false,"first_name":"","username":"","language_code":""},"language":"","custom_emoji_id":""}]}}`,
	}

	b := New("")
	b.Command("/test", func(ctx context.Context, update *types.Update) error {
		t.Log(update.Message.Args())

		return nil
	})

	for _, arg := range args {
		u := &types.Update{}
		err := json.Unmarshal([]byte(arg), &u)
		if err != nil {
			t.Fatal(err)
		}

		err = b.HandleUpdate(context.Background(), u)
		if err != nil {
			t.Fatal(err)
		}
	}
}
