package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Exchange struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
} 

func main() {
	// ctx, cancel := context.WithTimeout(context.Background(), 300 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 500 * time.Millisecond)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Print(err.Error())
	}
	
	response, err := http.DefaultClient.Do(request)

	if err != nil {
		msg := err.Error()

		// Verifica se o erro é do prazo do context que excedeu
		if strings.Contains(err.Error(), "context deadline exceeded") {
			msg = "context deadline exceed"
		}
		log.Print(msg)
	}

	if response != nil {
		defer response.Body.Close()
	
		body, err := io.ReadAll(response.Body)
		if err != nil {
			log.Print(err.Error())
		}

		var exchange Exchange
		err = json.Unmarshal(body, &exchange)

		if err != nil {
			log.Print(err.Error())
		}

		file, err := os.Create("cotacao.txt")
		if err != nil {
			log.Print(err.Error())
		}

		_, err = file.WriteString("Dolar: {"+exchange.Bid+"}" )
		if err != nil {
			log.Print(err.Error())
		}
		file.Close()
	}
}