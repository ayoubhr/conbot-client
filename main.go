package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	tele "gopkg.in/telebot.v3"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Resp struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Ratio  float64 `json:"exchange-rate"`
	Qty    float64 `json:"quantity"`
	Result float64 `json:"result"`
}

type List struct {
	Currencies []string
}

const (
	CONVERT    = "/convert"
	COMMANDS   = "/commands"
	HELP       = "/help"
	CURRENCIES = "/currencies"
)

func uriBuilder(quantity, fromCurrency, toCurrency string) string {
	// Building the uri with the received parameters
	return "http://localhost:8080/convert?from=" + fromCurrency + "&to=" + toCurrency + "&q=" + quantity
}

func convertService(uri string) Resp {

	// Building the request that will be sent to the converter service.
	req, _ := http.NewRequest("GET", uri, nil)

	// Perform the request to the converter service.
	res, _ := http.DefaultClient.Do(req)

	// Decodes the response value representing the result
	// into a variable of type Resp and returns it.
	var resp Resp
	err := json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		fmt.Println("conbot service is not available.")
	}

	return resp
}

func listService() ([]byte, error) {
	req, _ := http.NewRequest("GET", "https://currency-exchange.p.rapidapi.com/listquotes", nil)
	key := os.Getenv("RAPID_CURRENCY_KEY")
	host := os.Getenv("RAPID_CURRENCY_HOST")
	req.Header.Add("X-RapidAPI-Key", key)
	req.Header.Add("X-RapidAPI-Host", host)
	res, _ := http.DefaultClient.Do(req)
	return io.ReadAll(res.Body)
}

func requestsHandler(b *tele.Bot) {
	b.Handle(CONVERT, func(c tele.Context) error {
		if len(c.Data()) <= 0 {
			return c.Send("Please fill in your currencies for conversion or type\n---> /help command-name <--- for help.")
		}
		data := strings.Split(c.Data(), " ")
		log.Println("Client data: " + fmt.Sprintf("%v", data))

		uri := uriBuilder(data[0], strings.ToUpper(data[1]), strings.ToUpper(data[2]))
		log.Println("URI: " + uri)

		result := convertService(uri)
		log.Println("Result: " + fmt.Sprintf("%v", result))

		response := fmt.Sprintf("%.2f", result.Result) + " " + result.To
		return c.Send(response)
	})

	b.Handle(COMMANDS, func(c tele.Context) error {
		response := "List of commands available:\n" + CONVERT + "\n" + COMMANDS + "\n" + HELP + "\n" + CURRENCIES + "\n"
		return c.Send(response)
	})

	b.Handle(CURRENCIES, func(c tele.Context) error {
		result, _ := listService()
		fmt.Println(string(result))
		return c.Send(string(result))
	})
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalln("Environment variables could not load.")
	}

	pref := tele.Settings{
		Token:  os.Getenv("TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatalln(err)
		return
	}

	requestsHandler(b)

	b.Start()
}
