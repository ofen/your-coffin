package main

import (
	"context"
	"encoding/xml"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
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

func (m *meters) setHotWater(value int) {
	m.hotWater = value
}

func (m *meters) setColdWater(value int) {
	m.coldWater = value
}

func (m *meters) setElectricityT1(value int) {
	m.electricityT1 = value
}

func (m *meters) setElectricityT2(value int) {
	m.electricityT2 = value
}

type allowedUsers []struct {
	Username string `json:"username"`
}

type handlerFunc func(context.Context, *tgbotapi.Message)

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
