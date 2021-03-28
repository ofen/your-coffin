package main

import (
	"encoding/xml"
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
