package main

import (
	"encoding/xml"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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

type user struct {
	ID int64 `json:"id"`
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
