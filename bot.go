package main

import (
    "encoding/xml"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "strconv"
    "strings"
    "time"

    "golang.org/x/text/encoding/charmap"
    "github.com/go-telegram-bot-api/telegram-bot-api"
)

type ValCurs struct {
    XMLName xml.Name `xml:"ValCurs"`
    Text    string   `xml:",chardata"`
    Date    string   `xml:"Date,attr"`
    Name    string   `xml:"name,attr"`
    Valute  []struct {
        Text     string `xml:",chardata"`
        ID       string `xml:"ID,attr"`
        NumCode  string `xml:"NumCode"`
        CharCode string `xml:"CharCode"`
        Nominal  int `xml:"Nominal"`
        Name     string `xml:"Name"`
        Value    string `xml:"Value"`
    } `xml:"Valute"`
} 

type ExchangeRates struct {
    Date time.Time `xml:"Date,attr"`
    Name string `xml:"name,attr"`
    Rates []Rate `xml:"Valute"`
}

type Rate struct {
    Id string `xml:"ID,attr"`
    NumCode int `xml:"NumCode"`
    CharCode string `xml:"CharCode"`
    Nominal int `xml:"Nominal"`
    Name string `xml:"Name"`
    Value float32 `xml:"Value"`
}

func identReader(charset string, input io.Reader) (io.Reader, error) {
    if charset == "windows-1251" {
        return charmap.Windows1251.NewDecoder().Reader(input), nil
    }
    return nil, fmt.Errorf("unknown charset: %s", charset)
}

func main() {
    bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
    if err != nil {
        log.Fatalln(err)
    }

    bot.Debug = true

    log.Printf("Authorized on account %s", bot.Self.UserName)

    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60

    updates, err := bot.GetUpdatesChan(u)

    for update := range updates {
        var msg tgbotapi.MessageConfig
        if update.Message == nil {
            continue
        }

        if update.Message.IsCommand() {
            switch update.Message.Command() {
            case "status":
                msg = tgbotapi.NewMessage(update.Message.Chat.ID, "OK")
            case "currency":
                var valCurs ValCurs
                var messageText string

                resp, err := http.Get("http://www.cbr.ru/scripts/XML_daily.asp")
                if err != nil {
                    log.Fatalln(err)
                }
                defer resp.Body.Close()

                decoder := xml.NewDecoder(resp.Body)
                decoder.CharsetReader = identReader
                err = decoder.Decode(&valCurs)
                if err != nil {
                    log.Fatalln(err)
                }
                var report []string

                for _, valute := range valCurs.Valute {
                    switch valute.CharCode {
                    case
                        "USD",
                        "EUR":
                        valuteValue, err := strconv.ParseFloat(strings.Replace(valute.Value, ",", ".", 1), 64)
                        if err != nil {
                            log.Fatalln(err)
                        }
                        report = append(report, fmt.Sprintf("**%s:**%.2f", valute.CharCode, valuteValue))
                    }
                    
                }

                messageText = strings.Join(report, " ")

                if messageText == "" {
                    messageText = "No exchange rate found"
                }

                msg = tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
            default:
                msg = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Unsupported command: %q", update.Message.Command()))
                msg.ParseMode = "markdown"
                // msg.ReplyToMessageID = update.Message.MessageID
            }

            log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
            bot.Send(msg)
        }


    }
}
