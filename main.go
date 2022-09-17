package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	tele "gopkg.in/telebot.v3"
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

	b.Handle("/convert", func(c tele.Context) error {
		data := strings.Split(c.Data(), " ")
		log.Println("Client data: " + fmt.Sprintf("%v", data))

		uri := uriBuilder(data[0], strings.ToUpper(data[1]), strings.ToUpper(data[2]))
		log.Println("URI: " + uri)

		result := convertService(uri)
		log.Println("Result: " + fmt.Sprintf("%v", result))

		response := fmt.Sprintf("%.2f", result.Result) + " " + result.To
		return c.Send(response)
	})

	b.Start()
}
