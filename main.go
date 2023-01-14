package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/ofen/yourcoffin/internal/bot"
	"github.com/ofen/yourcoffin/internal/bot/types"
	"github.com/ofen/yourcoffin/internal/googlesheets"
)

var (
	b      = bot.New(os.Getenv("BOT_TOKEN"))
	gs     = googlesheets.New(os.Getenv("GOOGLE_SPREADSHEET"))
	secret = os.Getenv("SECRET_TOKEN")
)

func main() {
	b.Command("/status", statusHandler)
	b.Command("/help", helpHandler)
	b.Command("/lastmeters", lastmetersHandler)
	b.Command("/meters", metersHandler)
	b.Command("/test", testHandler)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	port := os.Getenv("PORT")
	log.Println("listening on :" + port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	if h := r.Header.Get(bot.HeaderSecretToken); h != secret {
		http.Error(w, "", http.StatusForbidden)

		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "", http.StatusMethodNotAllowed)

		return
	}

	defer r.Body.Close()

	update := &types.Update{}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)

		return
	}

	if err := b.HandleUpdate(context.Background(), update); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
