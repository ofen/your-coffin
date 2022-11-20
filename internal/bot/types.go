package bot

import (
	"encoding/xml"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Data struct {
	UpdateID int     `json:"update_id"`
	Message  Message `json:"message"`
}

type Response struct {
	Text   string `json:"text"`
	ChatID int    `json:"chat_id"`
}

type Message struct {
	MessageID int    `json:"message_id"`
	From      From   `json:"from"`
	Chat      Chat   `json:"chat"`
	Date      int    `json:"date"`
	Text      string `json:"text"`
}

type Chat struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
	Type      string `json:"type"`
}

type From struct {
	ID           int    `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
}

type valCurs struct {
	XMLName xml.Name `xml:"ValCurs"`
	Text    string   `xml:",chardata"`
	Date    string   `xml:"Date,attr"`
	Name    string   `xml:"name,attr"`
	Valute  []struct {
		Text     string `xml:",chardata"`
		ID       string `xml:"ID,attr"`
		NumCode  string `xml:"NumCode"`
		CharCode string `xml:"CharCode"`
		Nominal  int    `xml:"Nominal"`
		Name     string `xml:"Name"`
		Value    string `xml:"Value"`
	} `xml:"Valute"`
}

type meters struct {
	hotWater      int
	coldWater     int
	electricityT1 int
	electricityT2 int
}

type handlerFunc func(*tg.Message)

type contextKey string

func (c contextKey) String() string {
	return string(c)
}

// type exchangeRates struct {
// 	Date  time.Time `xml:"Date,attr"`
// 	Name  string    `xml:"name,attr"`
// 	Rates []rate    `xml:"Valute"`
// }

// type rate struct {
// 	ID       string  `xml:"ID,attr"`
// 	NumCode  int     `xml:"NumCode"`
// 	CharCode string  `xml:"CharCode"`
// 	Nominal  int     `xml:"Nominal"`
// 	Name     string  `xml:"Name"`
// 	Value    float32 `xml:"Value"`
// }
